package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/agentflare-ai/go-jsonschema"
)

// Graph operation schemas for reuse
var (
	// AddNodeSchema defines the schema for adding a node
	AddNodeSchema = &jsonschema.Schema{
		Type:        "object",
		Description: "Add a new node to the graph",
		Properties: map[string]*jsonschema.Schema{
			"labels": {
				Type:        "array",
				Description: "Node labels/types",
				Items:       &jsonschema.Schema{Type: "string"},
			},
			"properties": {
				Type:        "object",
				Description: "Node properties as key-value pairs",
			},
		},
	}

	// AddEdgeSchema defines the schema for adding an edge
	AddEdgeSchema = &jsonschema.Schema{
		Type:        "object",
		Description: "Add a new edge/relationship to the graph",
		Properties: map[string]*jsonschema.Schema{
			"start_node": {
				Type:        "integer",
				Description: "ID of the start node",
			},
			"end_node": {
				Type:        "integer",
				Description: "ID of the end node",
			},
			"type": {
				Type:        "string",
				Description: "Relationship type",
			},
			"properties": {
				Type:        "object",
				Description: "Edge properties as key-value pairs",
			},
		},
		Required: []string{"start_node", "end_node", "type"},
	}

	// UpdateNodeSchema defines the schema for updating a node
	UpdateNodeSchema = &jsonschema.Schema{
		Type:        "object",
		Description: "Update an existing node in the graph",
		Properties: map[string]*jsonschema.Schema{
			"id": {
				Type:        "integer",
				Description: "ID of the node to update",
			},
			"labels": {
				Type:        "array",
				Description: "New labels (replaces existing)",
				Items:       &jsonschema.Schema{Type: "string"},
			},
			"properties": {
				Type:        "object",
				Description: "Properties to update (merges with existing)",
			},
		},
		Required: []string{"id"},
	}

	// UpdateEdgeSchema defines the schema for updating an edge
	UpdateEdgeSchema = &jsonschema.Schema{
		Type:        "object",
		Description: "Update an existing edge in the graph",
		Properties: map[string]*jsonschema.Schema{
			"id": {
				Type:        "integer",
				Description: "ID of the edge to update",
			},
			"type": {
				Type:        "string",
				Description: "New relationship type (optional)",
			},
			"properties": {
				Type:        "object",
				Description: "Properties to update (merges with existing)",
			},
		},
		Required: []string{"id"},
	}

	// GraphUpdateSchema combines all operation schemas
	GraphUpdateSchema = &jsonschema.Schema{
		Type:        "object",
		Description: "A graph update operation",
		AnyOf: []*jsonschema.Schema{
			AddNodeSchema,
			AddEdgeSchema,
			UpdateNodeSchema,
			UpdateEdgeSchema,
		},
	}
)

// GraphDB represents a knowledge graph
type GraphDB struct {
	db          *sql.DB
	tableName   string
	nodesTable  string
	edgesTable  string
	vtAvailable bool
}

// Node represents a node in the graph
type Node struct {
	ID         int64          `json:"id"`
	Labels     []string       `json:"labels,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Relationship represents an edge/relationship in the graph
type Relationship struct {
	ID         int64          `json:"id"`
	StartNode  int64          `json:"start_node"`
	EndNode    int64          `json:"end_node"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Path represents a path through the graph
type Path struct {
	Nodes         []*Node         `json:"nodes"`
	Relationships []*Relationship `json:"relationships"`
}

// GraphUpdate represents an update operation for the graph
type GraphUpdate struct {
	Operation  string         `json:"operation"` // "create_node", "create_relationship", "update_node", "update_relationship"
	NodeData   *Node          `json:"node_data,omitempty"`
	RelData    *Relationship  `json:"rel_data,omitempty"`
	Labels     []string       `json:"labels,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// NewGraphDB creates a new graph instance with the given database and table name
func NewGraphDB(ctx context.Context, db *sql.DB, tableName string) (*GraphDB, error) {
	g := &GraphDB{
		db:        db,
		tableName: tableName,
	}

	if g.tableName == "" {
		g.tableName = "graph"
	}

	// Set backing table names
	g.nodesTable = g.tableName + "_nodes"
	g.edgesTable = g.tableName + "_edges"

	// Create backing tables with proper schema matching the C tests
	backingTablesQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s(
			id INTEGER PRIMARY KEY, 
			labels TEXT, 
			properties TEXT
		);
		CREATE TABLE IF NOT EXISTS %s(
			id INTEGER PRIMARY KEY, 
			source INTEGER, 
			target INTEGER, 
			edge_type TEXT, 
			weight REAL, 
			properties TEXT
		);
	`, g.nodesTable, g.edgesTable)

	if _, err := g.db.ExecContext(ctx, backingTablesQuery); err != nil {
		return nil, fmt.Errorf("failed to create backing tables: %w", err)
	}

	// Create the virtual table using the graph extension with backing tables
	// Check if the graph module is available first
	var moduleExists bool
	err := g.db.QueryRowContext(ctx, "SELECT 1 FROM pragma_module_list WHERE name = 'graph'").Scan(&moduleExists)
	if err == nil && moduleExists {
		query := fmt.Sprintf("CREATE VIRTUAL TABLE IF NOT EXISTS %s USING graph(%s, %s)",
			g.tableName, g.nodesTable, g.edgesTable)
		if _, err := g.db.ExecContext(ctx, query); err != nil {
			return nil, fmt.Errorf("failed to create graph table: %w", err)
		}
		g.vtAvailable = true
	}

	return g, nil
}

// CreateNode creates a new node with optional labels and properties
func (g *GraphDB) CreateNode(ctx context.Context, labels []string, properties map[string]any) (*Node, error) {
	// Prepare labels as JSON array
	labelsJSON := "[]"
	if len(labels) > 0 {
		if data, err := json.Marshal(labels); err == nil {
			labelsJSON = string(data)
		}
	}

	// Prepare properties as JSON string
	propertiesJSON := "{}"
	if len(properties) > 0 {
		if data, err := json.Marshal(properties); err == nil {
			propertiesJSON = string(data)
		}
	}

	if g.vtAvailable {
		// Insert through the virtual table interface
		query := fmt.Sprintf("INSERT INTO %s (type, labels, properties) VALUES (?, ?, ?)", g.tableName)
		result, err := g.db.ExecContext(ctx, query, "node", labelsJSON, propertiesJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to create node: %w", err)
		}
		// Get the node ID from LastInsertId
		nodeID, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get node ID: %w", err)
		}
		return &Node{ID: nodeID, Labels: labels, Properties: properties}, nil
	}
	// Fallback: insert directly into backing table
	query := fmt.Sprintf("INSERT INTO %s (labels, properties) VALUES (?, ?)", g.nodesTable)
	result, err := g.db.ExecContext(ctx, query, labelsJSON, propertiesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}
	nodeID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}

	return &Node{
		ID:         nodeID,
		Labels:     labels,
		Properties: properties,
	}, nil
}

// CreateRelationship creates a relationship between two nodes
func (g *GraphDB) CreateRelationship(ctx context.Context, startNodeID, endNodeID int64, relType string, properties map[string]interface{}) (*Relationship, error) {
	// Prepare properties as JSON string
	propertiesJSON := "{}"
	if len(properties) > 0 {
		if data, err := json.Marshal(properties); err == nil {
			propertiesJSON = string(data)
		}
	}

	if g.vtAvailable {
		// Insert through the virtual table interface with relationship validation
		query := fmt.Sprintf("INSERT INTO %s (type, from_id, to_id, rel_type, weight, properties) VALUES (?, ?, ?, ?, ?, ?)", g.tableName)
		result, err := g.db.ExecContext(ctx, query, "edge", startNodeID, endNodeID, relType, 1.0, propertiesJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to create relationship: %w", err)
		}
		// Get the relationship ID from LastInsertId (decode from virtual table rowid)
		relID, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get relationship ID: %w", err)
		}
		// Extract actual edge ID from encoded rowid (remove the edge bit flag)
		actualRelID := relID & ^(1 << 62)
		return &Relationship{ID: actualRelID, StartNode: startNodeID, EndNode: endNodeID, Type: relType, Properties: properties}, nil
	}
	// Fallback: insert directly into backing edges table, with manual validation
	// Validate start node
	var exists int
	if err := g.db.QueryRowContext(ctx, fmt.Sprintf("SELECT 1 FROM %s WHERE id=?", g.nodesTable), startNodeID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to create relationship: start node %d not found", startNodeID)
	}
	// Validate end node
	exists = 0
	if err := g.db.QueryRowContext(ctx, fmt.Sprintf("SELECT 1 FROM %s WHERE id=?", g.nodesTable), endNodeID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to create relationship: end node %d not found", endNodeID)
	}

	query := fmt.Sprintf("INSERT INTO %s (source, target, edge_type, weight, properties) VALUES (?, ?, ?, ?, ?)", g.edgesTable)
	result, err := g.db.ExecContext(ctx, query, startNodeID, endNodeID, relType, 1.0, propertiesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}
	relID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship ID: %w", err)
	}
	return &Relationship{ID: relID, StartNode: startNodeID, EndNode: endNodeID, Type: relType, Properties: properties}, nil
}

// FindNodes finds nodes matching the given criteria
func (g *GraphDB) FindNodes(ctx context.Context, labels []string, properties map[string]any) ([]*Node, error) {
	// WORKAROUND: Query backing table directly due to virtual table cursor bug
	// TODO: Switch back to virtual table once cursor is fixed
	query := fmt.Sprintf("SELECT id, labels, properties FROM %s", g.nodesTable)
	rows, err := g.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var id int64
		var labelsJSON, propertiesJSON string

		if err := rows.Scan(&id, &labelsJSON, &propertiesJSON); err != nil {
			continue
		}

		// Parse properties from JSON
		var nodeProps map[string]interface{}
		if propertiesJSON != "" {
			if err := json.Unmarshal([]byte(propertiesJSON), &nodeProps); err != nil {
				continue
			}
		}
		if nodeProps == nil {
			nodeProps = make(map[string]interface{})
		}

		// Parse labels from JSON array
		var nodeLabels []string
		if labelsJSON != "" && labelsJSON != "[]" {
			if err := json.Unmarshal([]byte(labelsJSON), &nodeLabels); err != nil {
				// Fallback to space-separated parsing if JSON fails
				nodeLabels = strings.Fields(labelsJSON)
			}
		}

		// Create node from data
		node := &Node{
			ID:         id,
			Labels:     nodeLabels,
			Properties: nodeProps,
		}

		// Filter by labels if specified
		if len(labels) > 0 {
			match := true
			for _, requiredLabel := range labels {
				found := false
				for _, nodeLabel := range node.Labels {
					if nodeLabel == requiredLabel {
						found = true
						break
					}
				}
				if !found {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		// Filter by properties if specified
		if len(properties) > 0 {
			match := true
			for k, v := range properties {
				if nodeVal, ok := node.Properties[k]; !ok || nodeVal != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// FindRelationships finds relationships matching the given criteria
func (g *GraphDB) FindRelationships(ctx context.Context, relType string, properties map[string]interface{}) ([]*Relationship, error) {
	// WORKAROUND: Query backing table directly due to virtual table cursor bug
	// TODO: Switch back to virtual table once cursor is fixed
	query := fmt.Sprintf("SELECT id, source, target, edge_type, weight, properties FROM %s", g.edgesTable)
	rows, err := g.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*Relationship
	for rows.Next() {
		var id, fromID, toID int64
		var edgeType, propertiesJSON string
		var weight float64

		if err := rows.Scan(&id, &fromID, &toID, &edgeType, &weight, &propertiesJSON); err != nil {
			continue
		}

		// Parse properties from JSON
		var allProps map[string]interface{}
		if propertiesJSON != "" {
			if err := json.Unmarshal([]byte(propertiesJSON), &allProps); err != nil {
				continue
			}
		}
		if allProps == nil {
			allProps = make(map[string]interface{})
		}

		// Use the edge_type column for the relationship type
		actualRelType := edgeType
		if actualRelType == "" {
			actualRelType = "UNKNOWN"
		}

		// Create relationship from data
		rel := &Relationship{
			ID:         id,
			StartNode:  fromID,
			EndNode:    toID,
			Type:       actualRelType,
			Properties: allProps,
		}

		// Filter by type if specified
		if relType != "" && rel.Type != relType {
			continue
		}

		// Filter by properties if specified
		if len(properties) > 0 {
			match := true
			for k, v := range properties {
				if relVal, ok := rel.Properties[k]; !ok || relVal != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		relationships = append(relationships, rel)
	}

	return relationships, nil
}

func (g *GraphDB) DeleteNode(ctx context.Context, nodeID int64) error {
	// Also delete relationships connected to this node
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", g.nodesTable)
	_, err := g.db.ExecContext(ctx, query, nodeID)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	query = fmt.Sprintf("DELETE FROM %s WHERE source = ? OR target = ?", g.edgesTable)
	_, err = g.db.ExecContext(ctx, query, nodeID, nodeID)
	if err != nil {
		return fmt.Errorf("failed to delete relationships for node: %w", err)
	}

	return nil
}

// Apply applies a GraphUpdate operation to the graph
func (g *GraphDB) Apply(ctx context.Context, update GraphUpdate) error {
	switch update.Operation {
	case "create_node":
		_, err := g.CreateNode(ctx, update.Labels, update.Properties)
		return err
	case "create_relationship":
		if update.RelData != nil {
			_, err := g.CreateRelationship(ctx, update.RelData.StartNode, update.RelData.EndNode, update.RelData.Type, update.RelData.Properties)
			return err
		}
		return fmt.Errorf("missing relationship data for create_relationship operation")
	default:
		return fmt.Errorf("unsupported operation: %s", update.Operation)
	}
}

// Search performs a search query on the graph and returns matching results as strings
func (g *GraphDB) Search(ctx context.Context, query string) ([]string, error) {
	// For now, perform a simple search on nodes
	// This is a placeholder implementation - in production you would want
	// more sophisticated graph traversal/search capabilities
	nodes, err := g.FindNodes(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, node := range nodes {
		// Simple text representation of the node
		nodeStr := fmt.Sprintf("Node[%d] Labels:%v Properties:%v", node.ID, node.Labels, node.Properties)
		results = append(results, nodeStr)
	}

	return results, nil
}

// Close closes the graph (does not close the underlying database connection)
func (g *GraphDB) Close() error {
	// Nothing to close for now, as we don't own the database connection
	return nil
}
