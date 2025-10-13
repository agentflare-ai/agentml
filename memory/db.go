//go:build !windows && cgo
// +build !windows,cgo

package memory

/*
#cgo CFLAGS: -DSQLITE_ENABLE_LOAD_EXTENSION=1
*/
import "C"

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/mattn/go-sqlite3"
)

type DB = sql.DB

// LoadGraphExtension loads the embedded graph extension into the database
func LoadGraphExtension(db *sql.DB) error {
	name := "graph_extension.so"
	if runtime.GOOS == "darwin" {
		name = "graph_extension.dylib"
	}
	return loadExtension(db, GraphExtension, name, "sqlite3_graph_init")
}

// LoadVecExtension loads the embedded vector extension into the database
func LoadVecExtension(db *sql.DB) error {
	name := "vec_extension.so"
	if runtime.GOOS == "darwin" {
		name = "vec_extension.dylib"
	}
	return loadExtension(db, VecExtension, name, "sqlite3_vec_init")
}

// loadExtension is a helper function that writes the embedded extension to a temporary file
// and loads it into the database
func loadExtension(db *sql.DB, extensionData []byte, filename, entryPoint string) error {
	// Create a temporary directory with a unique name
	tmpDir, err := os.MkdirTemp("", "sqlite-ext-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Ensure suffix is correct and remains at end of path
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]
	if runtime.GOOS == "darwin" {
		ext = ".dylib"
	}
	tmpFile := filepath.Join(tmpDir, base+ext)

	// Write the extension data to the temporary file with executable permissions
	if err := os.WriteFile(tmpFile, extensionData, 0755); err != nil {
		return fmt.Errorf("failed to write extension to temporary file: %w", err)
	}

	// On macOS, we might need to remove quarantine attributes
	if runtime.GOOS == "darwin" {
		// Try to remove quarantine attribute (this might fail but that's OK)
		exec.Command("xattr", "-d", "com.apple.quarantine", tmpFile).Run()
	}

	// Load the extension into the database
	loadPath := tmpFile
	if runtime.GOOS == "darwin" {
		ext := filepath.Ext(tmpFile)
		if ext == ".dylib" {
			loadPath = tmpFile[:len(tmpFile)-len(ext)]
		}
	}
	query := fmt.Sprintf("SELECT load_extension('%s', '%s')", loadPath, entryPoint)
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to load extension %s: %w", filename, err)
	}

	return nil
}

var registerOnce sync.Once

func NewDB(ctx context.Context, dsn string) (db *sql.DB, err error) {
	// Register the custom SQLite driver only once
	registerOnce.Do(func() {
		sql.Register("sqlite3_with_extensions", &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				// Enable extension loading first
				if _, err := conn.Exec("PRAGMA load_extension = 1", nil); err != nil {
					wrappedErr := fmt.Errorf("store.NewDB: failed to enable extension loading: %w", err)
					slog.Error(wrappedErr.Error())
					return wrappedErr
				}

				// Create temporary files for the extensions
				// Use platform-suffix and ensure filename ends with proper extension
				graphName := "graph_extension.so"
				if runtime.GOOS == "darwin" { graphName = "graph_extension-*.dylib" } else { graphName = "graph_extension-*.so" }
				graphTmpFile, err := writeExtensionToTemp(GraphExtension, graphName)
				if err != nil {
					wrappedErr := fmt.Errorf("store.NewDB: failed to write graph extension: %w", err)
					slog.Error(wrappedErr.Error())
					return wrappedErr
				}
				defer os.Remove(graphTmpFile)

				// On darwin, sqlite may append .dylib automatically; pass path without suffix
				graphLoadPath := graphTmpFile
				if runtime.GOOS == "darwin" {
					ext := filepath.Ext(graphTmpFile)
					if ext == ".dylib" {
						graphLoadPath = graphTmpFile[:len(graphTmpFile)-len(ext)]
					}
				}

				vecName := "vec_extension.so"
				if runtime.GOOS == "darwin" { vecName = "vec_extension-*.dylib" } else { vecName = "vec_extension-*.so" }
				vecTmpFile, err := writeExtensionToTemp(VecExtension, vecName)
				if err != nil {
					wrappedErr := fmt.Errorf("store.NewDB: failed to write vec extension: %w", err)
					slog.Error(wrappedErr.Error())
					return wrappedErr
				}
				defer os.Remove(vecTmpFile)

				// On darwin, sqlite may append .dylib automatically; pass path without suffix
				vecLoadPath := vecTmpFile
				if runtime.GOOS == "darwin" {
					ext := filepath.Ext(vecTmpFile)
					if ext == ".dylib" {
						vecLoadPath = vecTmpFile[:len(vecTmpFile)-len(ext)]
					}
				}

				// Load the extensions (best-effort). If unavailable on this platform, continue.
				if err := conn.LoadExtension(graphLoadPath, "sqlite3_graph_init"); err != nil {
					slog.Warn("store.NewDB: graph extension unavailable; continuing without it", "err", err)
				} else {
					slog.Debug("store.NewDB: graph extension loaded")
				}

				if err := conn.LoadExtension(vecLoadPath, "sqlite3_vec_init"); err != nil {
					slog.Warn("store.NewDB: vec extension unavailable; continuing without it", "err", err)
				} else {
					slog.Debug("store.NewDB: vec extension loaded")
				}

				return nil
			},
		})
	})

	// Open the database with the custom driver
	db, err = sql.Open("sqlite3_with_extensions", dsn)
	if err != nil {
		wrappedErr := fmt.Errorf("store.NewDB: failed to open database: %w", err)
		slog.Error(wrappedErr.Error())
		return nil, wrappedErr
	}
	if err := db.PingContext(ctx); err != nil {
		wrappedErr := fmt.Errorf("store.NewDB: failed to ping database: %w", err)
		slog.Error(wrappedErr.Error())
		return nil, wrappedErr
	}
	return db, nil
}

// writeExtensionToTemp writes extension data to a temporary file and returns the path
func writeExtensionToTemp(extensionData []byte, pattern string) (string, error) {
	// Ensure the created file ends with the correct platform suffix by using a pattern with '*'
	// Example: "graph_extension-*.dylib" or "vec_extension-*.so"
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if err := tmpFile.Chmod(0755); err != nil {
		return "", err
	}

	if _, err := tmpFile.Write(extensionData); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
