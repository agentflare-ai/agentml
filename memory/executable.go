package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"strings"
	"time"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DBTX abstracts *sql.DB and *sql.Tx for Exec/Query operations.
// It matches the subset of methods used by memory executables.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (d *Deps) dbtx() DBTX {
	if d != nil && d.tx != nil {
		return d.tx
	}
	return d.DB
}

// MemoryNamespaceURI is the XML namespace for memory executables.
const MemoryNamespaceURI = "github.com/agentflare-ai/agentml/memory"

// Deps holds dependencies for memory executables.
type Deps struct {
	DB          *sql.DB
	Graph       *GraphDB
	Vector      *VectorDB
	DefaultDims int
	// Embed computes the embedding for the provided text using the given model.
	Embed func(ctx context.Context, model, text string) ([]float32, error)
	// internal transaction (single-session convenience). Production code would track tx per store.
	tx *sql.Tx
}

// InitializeMemorySystem creates a fully initialized memory system with DB, Graph, and Vector stores.
// This is a convenience function that sets up everything needed for memory executables.
// dsn can be ":memory:" for in-memory database or a file path for persistent storage.
// vectorDims specifies the dimension for vector embeddings (default: 1536).
func InitializeMemorySystem(ctx context.Context, dsn string, vectorDims int) (*Deps, error) {
	if dsn == "" {
		dsn = ":memory:"
	}
	if vectorDims <= 0 {
		vectorDims = 1536 // Default for OpenAI embeddings
	}

	// Create the database
	db, err := NewDB(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Initialize graph store
	graph, err := NewGraphDB(ctx, db, "graph")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create graph store: %w", err)
	}

	// Initialize vector store
	vector, err := NewVectorDB(ctx, db, "vectors", vectorDims)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}

	// Ensure KV table exists
	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS kv(key TEXT PRIMARY KEY, value TEXT)")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create KV table: %w", err)
	}

	return &Deps{
		DB:          db,
		Graph:       graph,
		Vector:      vector,
		DefaultDims: vectorDims,
	}, nil
}

// ----------------------- Executables -----------------------

type memExec struct {
	xmldom.Element
	deps  *Deps
	local string
}

func (e *memExec) Execute(ctx context.Context, interp agentml.Interpreter) error {
	tr := otel.Tracer("memory")
	ctx, span := tr.Start(ctx, "memory."+e.local)
	defer span.End()

	dm := interp.DataModel()
	if dm == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "No data model available for memory",
			Data:      map[string]any{"element": e.local},
			Cause:     fmt.Errorf("no datamodel"),
		}
	}

	switch e.local {
	case "close":
		return e.execClose(ctx, dm)
	case "put":
		return e.execPut(ctx, dm)
	case "get":
		return e.execGet(ctx, dm)
	case "delete":
		return e.execDelete(ctx, dm)
	case "copy":
		return e.execCopy(ctx, dm)
	case "move":
		return e.execMove(ctx, dm)
	case "query":
		return e.execQuery(ctx, dm)
	case "kvtruncate":
		return e.execKVTruncate(ctx, dm)
	case "exec":
		return e.execSQL(ctx, dm) // exec is an alias for sql
	case "begin":
		return e.execBegin(ctx)
	case "commit":
		return e.execCommit(ctx)
	case "rollback":
		return e.execRollback(ctx)
	case "savepoint":
		return e.execSavepoint(ctx, dm)
	case "release":
		return e.execRelease(ctx, dm)
	case "sql":
		return e.execSQL(ctx, dm)
	case "embed":
		return e.execEmbed(ctx, dm)
	case "upsertvector":
		return e.execUpsertVector(ctx, dm)
	case "search":
		return e.execSearch(ctx, dm)
	case "deletevector":
		return e.execDeleteVector(ctx, dm)
	case "vectorindex":
		// auto-initialized on open; treat as success
		return nil
	case "addnode":
		return e.execAddNode(ctx, dm)
	case "addedge":
		return e.execAddEdge(ctx, dm)
	case "getnode":
		return e.execGetNode(ctx, dm)
	case "getedge":
		return e.execGetEdge(ctx, dm)
	case "deletenode":
		return e.execDeleteNode(ctx, dm)
	case "deleteedge":
		return e.execDeleteEdge(ctx, dm)
	case "neighbors", "getneighbors":
		return e.execNeighbors(ctx, dm)
	case "graphpath":
		return e.execGraphPath(ctx, dm)
	case "graphtruncate":
		return e.execGraphTruncate(ctx, dm)
	case "graphquery":
		return e.execGraphQuery(ctx, dm)
	default:
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "unsupported memory element",
			Data:      map[string]any{"element": e.local},
			Cause:     fmt.Errorf("unsupported"),
		}
	}
}

// ---- Core storage helpers (KV) ----

func (e *memExec) ensureKV(ctx context.Context) error {
	if e.deps == nil || e.deps.DB == nil {
		return fmt.Errorf("memory DB not configured")
	}
	_, err := e.deps.dbtx().ExecContext(ctx, "CREATE TABLE IF NOT EXISTS kv(key TEXT PRIMARY KEY, value TEXT)")
	return err
}

func (e *memExec) execClose(ctx context.Context, dm agentml.DataModel) error {
	if e.deps != nil {
		if e.deps.Graph != nil {
			_ = e.deps.Graph.Close()
			e.deps.Graph = nil
		}
		if e.deps.Vector != nil {
			_ = e.deps.Vector.Close()
			e.deps.Vector = nil
		}
		if e.deps.DB != nil {
			_ = e.deps.DB.Close()
			e.deps.DB = nil
		}
	}
	return nil
}

func (e *memExec) execPut(ctx context.Context, dm agentml.DataModel) error {
	if err := e.ensureKV(ctx); err != nil {
		return err
	}
	// Support both key and keyexpr
	key, err := getStringOrExpr(ctx, dm, e.Element, "key", "keyexpr")
	if err != nil {
		return err
	}
	if strings.TrimSpace(key) == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "missing key or keyexpr",
			Cause:     fmt.Errorf("missing key"),
		}
	}

	// Handle both value and valueexpr attributes
	var v any
	if valExpr := string(e.Element.GetAttribute("valueexpr")); valExpr != "" {
		v, err = dm.EvaluateValue(ctx, valExpr)
		if err != nil {
			return err
		}
	} else if val := string(e.Element.GetAttribute("value")); val != "" {
		// Use the literal string value
		v = val
	} else {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "missing value or valueexpr",
			Cause:     fmt.Errorf("missing value"),
		}
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = e.deps.dbtx().ExecContext(ctx, "INSERT INTO kv(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", key, string(data))
	return err
}

func (e *memExec) execGet(ctx context.Context, dm agentml.DataModel) error {
	if err := e.ensureKV(ctx); err != nil {
		return err
	}
	// Support both key and keyexpr
	key, err := getStringOrExpr(ctx, dm, e.Element, "key", "keyexpr")
	if err != nil {
		return err
	}
	// Handle both location and dataid attributes
	loc := string(e.Element.GetAttribute("location"))
	if loc == "" {
		loc = string(e.Element.GetAttribute("dataid"))
	}
	row := e.deps.dbtx().QueryRowContext(ctx, "SELECT value FROM kv WHERE key=?", key)
	var s string
	scanErr := row.Scan(&s)
	if scanErr != nil {
		assignIf(ctx, dm, loc, nil)
		return nil
	}
	var out any
	_ = json.Unmarshal([]byte(s), &out)
	assignIf(ctx, dm, loc, out)
	return nil
}

func (e *memExec) execDelete(ctx context.Context, dm agentml.DataModel) error {
	if err := e.ensureKV(ctx); err != nil {
		return err
	}
	// Support both key and keyexpr
	key, err := getStringOrExpr(ctx, dm, e.Element, "key", "keyexpr")
	if err != nil {
		return err
	}
	_, err = e.deps.dbtx().ExecContext(ctx, "DELETE FROM kv WHERE key=?", key)
	return err
}

func (e *memExec) execCopy(ctx context.Context, dm agentml.DataModel) error {
	if err := e.ensureKV(ctx); err != nil {
		return err
	}
	// Support both src/srcexpr and srckey/srckeyexpr
	srcKey, err := getStringOrExpr(ctx, dm, e.Element, "src", "srcexpr")
	if err != nil {
		return err
	}
	if srcKey == "" {
		srcKey, err = getStringOrExpr(ctx, dm, e.Element, "srckey", "srckeyexpr")
		if err != nil {
			return err
		}
	}
	// Support both dst/dstexpr and dstkey/dstkeyexpr
	dstKey, err := getStringOrExpr(ctx, dm, e.Element, "dst", "dstexpr")
	if err != nil {
		return err
	}
	if dstKey == "" {
		dstKey, err = getStringOrExpr(ctx, dm, e.Element, "dstkey", "dstkeyexpr")
		if err != nil {
			return err
		}
	}
	if srcKey == "" || dstKey == "" {
		return fmt.Errorf("copy requires src/srckey and dst/dstkey attributes")
	}
	_, err = e.deps.dbtx().ExecContext(ctx,
		"INSERT INTO kv(key,value) SELECT ?, value FROM kv WHERE key=? ON CONFLICT(key) DO UPDATE SET value=excluded.value",
		dstKey, srcKey)
	return err
}

func (e *memExec) execMove(ctx context.Context, dm agentml.DataModel) error {
	if err := e.ensureKV(ctx); err != nil {
		return err
	}
	// Support both src/srcexpr and srckey/srckeyexpr
	srcKey, err := getStringOrExpr(ctx, dm, e.Element, "src", "srcexpr")
	if err != nil {
		return err
	}
	if srcKey == "" {
		srcKey, err = getStringOrExpr(ctx, dm, e.Element, "srckey", "srckeyexpr")
		if err != nil {
			return err
		}
	}
	// Support both dst/dstexpr and dstkey/dstkeyexpr
	dstKey, err := getStringOrExpr(ctx, dm, e.Element, "dst", "dstexpr")
	if err != nil {
		return err
	}
	if dstKey == "" {
		dstKey, err = getStringOrExpr(ctx, dm, e.Element, "dstkey", "dstkeyexpr")
		if err != nil {
			return err
		}
	}
	if srcKey == "" || dstKey == "" {
		return fmt.Errorf("move requires src/srckey and dst/dstkey attributes")
	}
	// If a transaction is active, use it; otherwise create a short-lived transaction for atomicity
	if e.deps != nil && e.deps.tx != nil {
		// Copy then delete using active tx
		if _, err := e.deps.tx.ExecContext(ctx,
			"INSERT INTO kv(key,value) SELECT ?, value FROM kv WHERE key=? ON CONFLICT(key) DO UPDATE SET value=excluded.value",
			dstKey, srcKey); err != nil {
			return err
		}
		if _, err := e.deps.tx.ExecContext(ctx, "DELETE FROM kv WHERE key=?", srcKey); err != nil {
			return err
		}
		return nil
	}
	// No active tx: use a local transaction on the base DB
	tx, err := e.deps.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO kv(key,value) SELECT ?, value FROM kv WHERE key=? ON CONFLICT(key) DO UPDATE SET value=excluded.value",
		dstKey, srcKey); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM kv WHERE key=?", srcKey); err != nil {
		return err
	}
	return tx.Commit()
}

func (e *memExec) execQuery(ctx context.Context, dm agentml.DataModel) error {
	// Query is similar to SQL but for KV operations
	// Support both sql/sqlexpr and query/queryexpr
	sql, err := getStringOrExpr(ctx, dm, e.Element, "sql", "sqlexpr")
	if err != nil {
		return err
	}
	if sql == "" {
		sql, err = getStringOrExpr(ctx, dm, e.Element, "query", "queryexpr")
		if err != nil {
			return err
		}
	}
	loc := string(e.Element.GetAttribute("location"))
	if loc == "" {
		loc = string(e.Element.GetAttribute("dataid"))
	}
	if sql == "" {
		return fmt.Errorf("query requires sql/query attribute")
	}
	rows, err := e.deps.dbtx().QueryContext(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var out []map[string]any
	for rows.Next() {
		scan := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range scan {
			ptrs[i] = &scan[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		m := map[string]any{}
		for i, c := range cols {
			m[c] = scan[i]
		}
		out = append(out, m)
	}
	assignIf(ctx, dm, loc, out)
	return nil
}

func (e *memExec) execKVTruncate(ctx context.Context, dm agentml.DataModel) error {
	if err := e.ensureKV(ctx); err != nil {
		return err
	}
	_, err := e.deps.dbtx().ExecContext(ctx, "DELETE FROM kv")
	return err
}

func (e *memExec) execBegin(ctx context.Context) error {
	if e.deps == nil || e.deps.DB == nil {
		return fmt.Errorf("memory DB not configured")
	}
	if e.deps.tx != nil {
		return nil
	}
	tx, err := e.deps.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	e.deps.tx = tx
	return nil
}

func (e *memExec) execCommit(ctx context.Context) error {
	if e.deps == nil || e.deps.tx == nil {
		return nil
	}
	err := e.deps.tx.Commit()
	e.deps.tx = nil
	return err
}

func (e *memExec) execRollback(ctx context.Context) error {
	if e.deps == nil || e.deps.tx == nil {
		return nil
	}
	err := e.deps.tx.Rollback()
	e.deps.tx = nil
	return err
}

func (e *memExec) execSavepoint(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.DB == nil {
		return fmt.Errorf("memory DB not configured")
	}
	// Support both name and nameexpr
	name, err := getStringOrExpr(ctx, dm, e.Element, "name", "nameexpr")
	if err != nil {
		return err
	}
	if name == "" {
		name = "sp_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	_, err = e.deps.dbtx().ExecContext(ctx, "SAVEPOINT "+name)
	return err
}

func (e *memExec) execRelease(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.DB == nil {
		return fmt.Errorf("memory DB not configured")
	}
	// Support both name and nameexpr
	name, err := getStringOrExpr(ctx, dm, e.Element, "name", "nameexpr")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("release requires name or nameexpr attribute")
	}
	_, err = e.deps.dbtx().ExecContext(ctx, "RELEASE SAVEPOINT "+name)
	return err
}

func (e *memExec) execSQL(ctx context.Context, dm agentml.DataModel) error {
	// Handle both sql and sqlexpr attributes
	sql := string(e.Element.GetAttribute("sql"))
	if sql == "" {
		sqlexpr := string(e.Element.GetAttribute("sqlexpr"))
		if sqlexpr == "" {
			return fmt.Errorf("exec requires sql or sqlexpr attribute")
		}
		var err error
		sql, err = evalString(ctx, dm, sqlexpr)
		if err != nil {
			return err
		}
	}
	// Handle both location and dataid attributes
	loc := string(e.Element.GetAttribute("location"))
	if loc == "" {
		loc = string(e.Element.GetAttribute("dataid"))
	}
	// If there's no location/dataid specified, it's an exec, not a query
	if loc == "" {
		// Execute the SQL statement (CREATE TABLE, INSERT, etc.)
		_, err := e.deps.dbtx().ExecContext(ctx, sql)
		return err
	} else {
		// Query and store results
		rows, err := e.deps.dbtx().QueryContext(ctx, sql)
		if err != nil {
			return err
		}
		defer rows.Close()
		cols, _ := rows.Columns()
		var out []map[string]any
		for rows.Next() {
			scan := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range scan {
				ptrs[i] = &scan[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				continue
			}
			m := map[string]any{}
			for i, c := range cols {
				m[c] = scan[i]
			}
			out = append(out, m)
		}
		assignIf(ctx, dm, loc, out)
		return nil
	}
}

// ---- Embeddings & Vectors ----

func (e *memExec) execEmbed(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Embed == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "embedder_unavailable",
			Cause:     fmt.Errorf("no embedder"),
		}
	}
	// Support both model and modelexpr
	model, err := getStringOrExpr(ctx, dm, e.Element, "model", "modelexpr")
	if err != nil {
		return err
	}
	// Support both text and textexpr
	text, err := getStringOrExpr(ctx, dm, e.Element, "text", "textexpr")
	if err != nil {
		return err
	}
	loc := string(e.Element.GetAttribute("location"))
	// Support both key and keyexpr
	key, err := getStringOrExpr(ctx, dm, e.Element, "key", "keyexpr")
	if err != nil {
		return err
	}
	vec, err := e.deps.Embed(ctx, model, text)
	if err != nil {
		return err
	}
	assignIf(ctx, dm, loc, vec)
	// Optional upsert
	if key != "" && e.deps.Vector != nil {
		id := hashKey(key)
		if err := e.deps.Vector.InsertVector(ctx, id, vec); err != nil {
			return err
		}
	}
	return nil
}

func (e *memExec) execUpsertVector(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Vector == nil {
		return fmt.Errorf("vector store not configured")
	}
	// Support both key and keyexpr
	key, err := getStringOrExpr(ctx, dm, e.Element, "key", "keyexpr")
	if err != nil {
		return err
	}
	// Support both vector and vectorexpr
	var v any
	if vecExpr := string(e.Element.GetAttribute("vectorexpr")); vecExpr != "" {
		v, err = dm.EvaluateValue(ctx, vecExpr)
		if err != nil {
			return err
		}
	} else if vec := string(e.Element.GetAttribute("vector")); vec != "" {
		v, err = dm.EvaluateValue(ctx, vec)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("upsertvector requires vector or vectorexpr")
	}
	arr, ok := toFloat32Slice(v)
	if !ok {
		return fmt.Errorf("vector must evaluate to []number")
	}
	return e.deps.Vector.InsertVector(ctx, hashKey(key), arr)
}

func (e *memExec) execSearch(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Vector == nil || e.deps.Embed == nil {
		return fmt.Errorf("vector search not available")
	}
	// Support both model and modelexpr
	model, err := getStringOrExpr(ctx, dm, e.Element, "model", "modelexpr")
	if err != nil {
		return err
	}
	// Support both text and textexpr
	text, err := getStringOrExpr(ctx, dm, e.Element, "text", "textexpr")
	if err != nil {
		return err
	}
	loc := string(e.Element.GetAttribute("location"))
	// Support both topk and topkexpr
	topkStr, err := getStringOrExpr(ctx, dm, e.Element, "topk", "topkexpr")
	if err != nil {
		return err
	}
	topk := 5
	if strings.TrimSpace(topkStr) != "" {
		fmt.Sscan(topkStr, &topk)
	}
	qvec, err := e.deps.Embed(ctx, model, text)
	if err != nil {
		return err
	}
	res, err := e.deps.Vector.SearchSimilarVectors(ctx, qvec, topk)
	if err != nil {
		return err
	}
	// Convert to array of maps
	outs := make([]map[string]any, 0, len(res))
	for _, r := range res {
		outs = append(outs, map[string]any{"id": r.ID, "distance": r.Distance})
	}
	assignIf(ctx, dm, loc, outs)
	return nil
}

func (e *memExec) execDeleteVector(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Vector == nil {
		return fmt.Errorf("vector store not configured")
	}
	// Support both key and keyexpr
	key, err := getStringOrExpr(ctx, dm, e.Element, "key", "keyexpr")
	if err != nil {
		return err
	}
	_, err = e.deps.dbtx().ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE rowid=?", e.deps.Vector.tableName), int64(hashKey(key)))
	return err
}

// ---- Graph helpers ----

func (e *memExec) execAddNode(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph not configured")
	}
	// Support both labels and labelsexpr
	labelsStr, err := getStringOrExpr(ctx, dm, e.Element, "labels", "labelsexpr")
	if err != nil {
		return err
	}
	labels := parseLabels(labelsStr)
	// Support both props and propsexpr
	var props map[string]any
	if propsExpr := string(e.Element.GetAttribute("propsexpr")); propsExpr != "" {
		props, _ = evalMap(ctx, dm, propsExpr)
	} else if propsVal := string(e.Element.GetAttribute("props")); propsVal != "" {
		props, _ = evalMap(ctx, dm, propsVal)
	}
	n, err := e.deps.Graph.CreateNode(ctx, labels, props)
	if err != nil {
		return err
	}
	assignIf(ctx, dm, string(e.Element.GetAttribute("location")), n)
	return nil
}

func (e *memExec) execAddEdge(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph not configured")
	}
	// Support both src and srcexpr
	src, err := getIntOrExpr(ctx, dm, e.Element, "src", "srcexpr")
	if err != nil {
		return err
	}
	// Support both dst and dstexpr
	dst, err := getIntOrExpr(ctx, dm, e.Element, "dst", "dstexpr")
	if err != nil {
		return err
	}
	// Support both rel and relexpr
	rel, err := getStringOrExpr(ctx, dm, e.Element, "rel", "relexpr")
	if err != nil {
		return err
	}
	// Support both props and propsexpr
	var props map[string]any
	if propsExpr := string(e.Element.GetAttribute("propsexpr")); propsExpr != "" {
		props, _ = evalMap(ctx, dm, propsExpr)
	} else if propsVal := string(e.Element.GetAttribute("props")); propsVal != "" {
		props, _ = evalMap(ctx, dm, propsVal)
	}
	slog.InfoContext(ctx, "memory: adding edge", "src", src, "dst", dst, "rel", rel)
	_, err = e.deps.Graph.CreateRelationship(ctx, src, dst, rel, props)
	if err != nil {
		slog.WarnContext(ctx, "memory: failed to add edge", "error", err)
		return err
	}
	slog.InfoContext(ctx, "memory: edge added")
	return nil
}

func (e *memExec) execGetNode(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph not configured")
	}
	// Support both id and idexpr
	id, err := getIntOrExpr(ctx, dm, e.Element, "id", "idexpr")
	if err != nil {
		return err
	}
	loc := string(e.Element.GetAttribute("location"))
	row := e.deps.dbtx().QueryRowContext(ctx, fmt.Sprintf("SELECT id, labels, properties FROM %s WHERE id=?", e.deps.Graph.nodesTable), id)
	var nid int64
	var labelsJSON, propsJSON string
	if err := row.Scan(&nid, &labelsJSON, &propsJSON); err != nil {
		assignIf(ctx, dm, loc, nil)
		return nil
	}
	var labels []string
	var props map[string]any
	_ = json.Unmarshal([]byte(labelsJSON), &labels)
	_ = json.Unmarshal([]byte(propsJSON), &props)
	assignIf(ctx, dm, loc, &Node{ID: nid, Labels: labels, Properties: props})
	return nil
}

func (e *memExec) execDeleteNode(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph not configured")
	}
	// Support both id and idexpr
	id, err := getIntOrExpr(ctx, dm, e.Element, "id", "idexpr")
	if err != nil {
		return err
	}
	return e.deps.Graph.DeleteNode(ctx, id)
}

func (e *memExec) execDeleteEdge(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph not configured")
	}
	// Support both src and srcexpr
	src, err := getIntOrExpr(ctx, dm, e.Element, "src", "srcexpr")
	if err != nil {
		return err
	}
	// Support both dst and dstexpr
	dst, err := getIntOrExpr(ctx, dm, e.Element, "dst", "dstexpr")
	if err != nil {
		return err
	}
	// Support both rel and relexpr
	rel, err := getStringOrExpr(ctx, dm, e.Element, "rel", "relexpr")
	if err != nil {
		return err
	}
	_, err = e.deps.dbtx().ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE source=? AND target=? AND edge_type=?", e.deps.Graph.edgesTable), src, dst, rel)
	return err
}

func (e *memExec) execNeighbors(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph not configured")
	}
	// Support both id and idexpr
	id, err := getIntOrExpr(ctx, dm, e.Element, "id", "idexpr")
	if err != nil {
		return err
	}
	// Support both direction and directionexpr
	dir, err := getStringOrExpr(ctx, dm, e.Element, "direction", "directionexpr")
	if err != nil {
		return err
	}
	dir = strings.ToLower(dir)
	loc := string(e.Element.GetAttribute("location"))
	var rows *sql.Rows
	if dir == "in" {
		rows, err = e.deps.dbtx().QueryContext(ctx, fmt.Sprintf("SELECT source FROM %s WHERE target=?", e.deps.Graph.edgesTable), id)
	} else {
		rows, err = e.deps.dbtx().QueryContext(ctx, fmt.Sprintf("SELECT target FROM %s WHERE source=?", e.deps.Graph.edgesTable), id)
	}
	if err != nil {
		return err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var nid int64
		if err := rows.Scan(&nid); err == nil {
			out = append(out, nid)
		}
	}
	slog.InfoContext(ctx, "memory: neighbors computed", "count", len(out), "location", loc)
	assignIf(ctx, dm, loc, out)
	return nil
}

func (e *memExec) execGetEdge(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph database not configured")
	}
	// Support both id and idexpr
	id, err := getStringOrExpr(ctx, dm, e.Element, "id", "idexpr")
	if err != nil {
		return err
	}
	loc := string(e.Element.GetAttribute("location"))
	if loc == "" {
		loc = string(e.Element.GetAttribute("dataid"))
	}
	// For now, use a simple query to get edge by ID
	rows, err := e.deps.dbtx().QueryContext(ctx,
		"SELECT id, source, target, edge_type, properties FROM relationships WHERE id = ?", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		var edge struct {
			ID         string
			Src        int64
			Dst        int64
			Rel        string
			Properties string
		}
		err = rows.Scan(&edge.ID, &edge.Src, &edge.Dst, &edge.Rel, &edge.Properties)
		if err != nil {
			return err
		}
		var props map[string]interface{}
		_ = json.Unmarshal([]byte(edge.Properties), &props)
		assignIf(ctx, dm, loc, map[string]interface{}{
			"id":         edge.ID,
			"src":        edge.Src,
			"dst":        edge.Dst,
			"type":       edge.Rel,
			"properties": props,
		})
	} else {
		assignIf(ctx, dm, loc, nil)
	}
	return nil
}

func (e *memExec) execGraphPath(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph database not configured")
	}
	// Support both src and srcexpr
	src, err := getIntOrExpr(ctx, dm, e.Element, "src", "srcexpr")
	if err != nil {
		return err
	}
	// Support both dst and dstexpr
	dst, err := getIntOrExpr(ctx, dm, e.Element, "dst", "dstexpr")
	if err != nil {
		return err
	}
	loc := string(e.Element.GetAttribute("location"))
	if loc == "" {
		loc = string(e.Element.GetAttribute("dataid"))
	}

	// Simple BFS path finding
	type pathNode struct {
		id   int64
		path []int64
	}

	visited := make(map[int64]bool)
	queue := []pathNode{{id: src, path: []int64{src}}}
	visited[src] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.id == dst {
			assignIf(ctx, dm, loc, current.path)
			return nil
		}

		// Get neighbors
		rows, err := e.deps.dbtx().QueryContext(ctx,
			fmt.Sprintf("SELECT target FROM %s WHERE source=?", e.deps.Graph.edgesTable), current.id)
		if err != nil {
			continue
		}

		for rows.Next() {
			var neighborID int64
			if err := rows.Scan(&neighborID); err == nil && !visited[neighborID] {
				visited[neighborID] = true
				newPath := append([]int64{}, current.path...)
				newPath = append(newPath, neighborID)
				queue = append(queue, pathNode{id: neighborID, path: newPath})
			}
		}
		rows.Close()
	}

	assignIf(ctx, dm, loc, nil) // No path found
	return nil
}

func (e *memExec) execGraphTruncate(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph database not configured")
	}
	// Clear both nodes and relationships tables
	_, err := e.deps.DB.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", e.deps.Graph.edgesTable))
	if err != nil {
		return err
	}
	_, err = e.deps.DB.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", e.deps.Graph.nodesTable))
	return err
}

func (e *memExec) execGraphQuery(ctx context.Context, dm agentml.DataModel) error {
	if e.deps == nil || e.deps.Graph == nil {
		return fmt.Errorf("graph not configured")
	}
	q := mustEvalString(ctx, dm, string(e.Element.GetAttribute("pathexpr")))
	res, err := e.deps.Graph.Search(ctx, q)
	if err != nil {
		return err
	}
	assignIf(ctx, dm, string(e.Element.GetAttribute("location")), res)
	return nil
}

// ---- Legacy graph: <memory:graph op="..."> ----

type graphExec struct {
	xmldom.Element
	deps      *Deps
	Op        string
	Out       string
	Labels    []string
	PropsExpr string
	StartExpr string
	EndExpr   string
	RelType   string
	IDExpr    string
	QueryExpr string
}

func (g *graphExec) Execute(ctx context.Context, interp agentml.Interpreter) error {
	tr := otel.Tracer("memory")
	ctx, span := tr.Start(ctx, "memory.graph.execute",
		trace.WithAttributes(attribute.String("op", g.Op)),
	)
	defer span.End()

	if g.deps == nil || g.deps.Graph == nil {
		err := fmt.Errorf("memory graph dependency not configured")
		span.SetStatus(codes.Error, err.Error())
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "memory graph dependency not configured",
			Data: map[string]any{
				"element": "memory:graph",
				"op":      g.Op,
			},
			Cause: err,
		}
	}

	dm := interp.DataModel()
	if dm == nil {
		err := fmt.Errorf("no data model available")
		span.SetStatus(codes.Error, err.Error())
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "No data model available for memory:graph",
			Data: map[string]any{
				"element": "memory:graph",
				"op":      g.Op,
			},
			Cause: err,
		}
	}

	// Parse attributes once here for legacy
	g.Op = strings.ToLower(string(g.Element.GetAttribute("op")))
	g.Out = string(g.Element.GetAttribute("out"))
	g.Labels = parseLabels(string(g.Element.GetAttribute("labels")))
	g.PropsExpr = string(g.Element.GetAttribute("properties-expr"))
	g.StartExpr = string(g.Element.GetAttribute("start"))
	g.EndExpr = string(g.Element.GetAttribute("end"))
	g.RelType = string(g.Element.GetAttribute("type"))
	g.IDExpr = string(g.Element.GetAttribute("id"))
	g.QueryExpr = string(g.Element.GetAttribute("query-expr"))
	if g.Op == "" {
		return fmt.Errorf("memory:graph requires op")
	}

	switch g.Op {
	case "create_node", "add_node", "addnode", "create-node":
		props, _ := evalMap(ctx, dm, g.PropsExpr)
		n, err := g.deps.Graph.CreateNode(ctx, g.Labels, props)
		if err != nil {
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   fmt.Sprintf("Failed to create node: %v", err),
				Data: map[string]any{
					"element": "memory:graph",
					"op":      g.Op,
				},
				Cause: err,
			}
		}
		assignIf(ctx, dm, g.Out, n)
		return nil
	case "create_relationship", "create_edge", "add_relationship", "add_edge", "create-relationship", "create-edge":
		startID, _ := evalInt64(ctx, dm, g.StartExpr)
		endID, _ := evalInt64(ctx, dm, g.EndExpr)
		props, _ := evalMap(ctx, dm, g.PropsExpr)
		_, err := g.deps.Graph.CreateRelationship(ctx, startID, endID, g.RelType, props)
		return err
	case "find_nodes", "find-nodes":
		props, _ := evalMap(ctx, dm, g.PropsExpr)
		nodes, err := g.deps.Graph.FindNodes(ctx, g.Labels, props)
		if err != nil {
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   fmt.Sprintf("Failed to find nodes: %v", err),
				Data: map[string]any{
					"element": "memory:graph",
					"op":      g.Op,
				},
				Cause: err,
			}
		}
		assignIf(ctx, dm, g.Out, nodes)
		return nil
	case "delete_node", "delete-node":
		id, _ := evalInt64(ctx, dm, g.IDExpr)
		return g.deps.Graph.DeleteNode(ctx, id)
	case "search", "graph-search":
		query, _ := evalString(ctx, dm, g.QueryExpr)
		results, err := g.deps.Graph.Search(ctx, query)
		if err != nil {
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   fmt.Sprintf("Failed to search graph: %v", err),
				Data: map[string]any{
					"element": "memory:graph",
					"op":      g.Op,
				},
				Cause: err,
			}
		}
		assignIf(ctx, dm, g.Out, results)
		return nil
	default:
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   fmt.Sprintf("unsupported memory:graph op '%s'", g.Op),
			Data: map[string]any{
				"element": "memory:graph",
				"op":      g.Op,
			},
			Cause: fmt.Errorf("unsupported"),
		}
	}
}

// ----------------------- Helpers -----------------------

func parseLabels(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func evalMap(ctx context.Context, dm agentml.DataModel, expr string) (map[string]any, error) {
	if strings.TrimSpace(expr) == "" {
		return nil, nil
	}
	v, err := dm.EvaluateValue(ctx, expr)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object for properties-expr, got %T", v)
	}
	return m, nil
}

func evalInt64(ctx context.Context, dm agentml.DataModel, expr string) (int64, error) {
	v, err := dm.EvaluateValue(ctx, expr)
	if err != nil {
		return 0, err
	}
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case float64:
		return int64(n), nil
	case string:
		var out int64
		_, err := fmt.Sscan(n, &out)
		if err != nil {
			return 0, fmt.Errorf("cannot parse id: %v", err)
		}
		return out, nil
	default:
		return 0, fmt.Errorf("unsupported id type: %T", v)
	}
}

func evalString(ctx context.Context, dm agentml.DataModel, expr string) (string, error) {
	if strings.TrimSpace(expr) == "" {
		return "", nil
	}
	v, err := dm.EvaluateValue(ctx, expr)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", v)
	}
	return s, nil
}

func mustEvalString(ctx context.Context, dm agentml.DataModel, expr string) string {
	s, _ := evalString(ctx, dm, expr)
	return s
}

// getStringOrExpr retrieves a string from either a literal attribute or an expression attribute.
// It checks attrExpr first (e.g., "keyexpr"), then falls back to attr (e.g., "key").
func getStringOrExpr(ctx context.Context, dm agentml.DataModel, el xmldom.Element, attr, attrExpr string) (string, error) {
	// Try expression attribute first
	if exprVal := string(el.GetAttribute(xmldom.DOMString(attrExpr))); exprVal != "" {
		return evalString(ctx, dm, exprVal)
	}
	// Fall back to literal attribute
	return string(el.GetAttribute(xmldom.DOMString(attr))), nil
}

// getIntOrExpr retrieves an int64 from either a literal attribute or an expression attribute.
func getIntOrExpr(ctx context.Context, dm agentml.DataModel, el xmldom.Element, attr, attrExpr string) (int64, error) {
	// Try expression attribute first
	if exprVal := string(el.GetAttribute(xmldom.DOMString(attrExpr))); exprVal != "" {
		return evalInt64(ctx, dm, exprVal)
	}
	// Fall back to literal attribute - need to evaluate if it's a number string
	if litVal := string(el.GetAttribute(xmldom.DOMString(attr))); litVal != "" {
		return evalInt64(ctx, dm, litVal)
	}
	return 0, nil
}

func toFloat32Slice(v any) ([]float32, bool) {
	switch arr := v.(type) {
	case []any:
		out := make([]float32, len(arr))
		for i, x := range arr {
			switch n := x.(type) {
			case float64:
				out[i] = float32(n)
			case int:
				out[i] = float32(n)
			default:
				return nil, false
			}
		}
		return out, true
	case []float64:
		out := make([]float32, len(arr))
		for i, n := range arr {
			out[i] = float32(n)
		}
		return out, true
	default:
		return nil, false
	}
}

func assignIf(ctx context.Context, dm agentml.DataModel, loc string, value any) {
	if strings.TrimSpace(loc) == "" {
		return
	}
	// Use SetVariable to avoid JS serialization issues for complex Go values (maps, slices, structs)
	if err := dm.SetVariable(ctx, loc, value); err != nil {
		slog.WarnContext(ctx, "memory: failed to assign result", "location", loc, "error", err)
	}
}

func raise(ctx context.Context, interp agentml.Interpreter, name string, data map[string]any) {
	if name == "" {
		return
	}
	evt := &agentml.Event{
		Name: name,
		Data: data,
		Type: agentml.EventTypeInternal,
	}
	interp.Raise(ctx, evt)
}

func hashKey(key string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return h.Sum64()
}
