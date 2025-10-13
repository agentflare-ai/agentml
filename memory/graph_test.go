package memory

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGraphVirtualTableIntegration(t *testing.T) {
	// Create an in-memory SQLite database
	ctx := context.Background()
	db, err := NewDB(ctx, ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if graph extension is actually available
	var moduleExists bool
	err = db.QueryRow("SELECT 1 FROM pragma_module_list WHERE name = 'graph'").Scan(&moduleExists)
	if err != nil {
		t.Logf("Graph extension not available: %v", err)
	} else {
		t.Logf("Graph extension IS available")
	}

	// Extension should already be loaded by NewDB, no need to load it again

	// Create a new graph instance
	graph, err := NewGraphDB(ctx, db, "test_graph")
	if err != nil {
		t.Fatalf("Failed to create graph: %v", err)
	}
	defer graph.Close()

	// Test node creation
	t.Run("CreateNode", func(t *testing.T) {
		node, err := graph.CreateNode(ctx, []string{"Person", "User"}, map[string]interface{}{
			"name": "Alice",
			"age":  30,
		})
		if err != nil {
			t.Fatalf("Failed to create node: %v", err)
		}

		if node.ID <= 0 {
			t.Error("Expected positive node ID")
		}
		if len(node.Labels) != 2 || node.Labels[0] != "Person" || node.Labels[1] != "User" {
			t.Errorf("Expected labels [Person, User], got %v", node.Labels)
		}
		if node.Properties["name"] != "Alice" || node.Properties["age"] != int(30) {
			t.Errorf("Expected properties {name: Alice, age: 30}, got %v", node.Properties)
		}
	})

	// Create test nodes for relationships
	node1, err := graph.CreateNode(ctx, []string{"Person"}, map[string]interface{}{
		"name": "Bob",
	})
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}

	node2, err := graph.CreateNode(ctx, []string{"Person"}, map[string]interface{}{
		"name": "Charlie",
	})
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}

	// Test relationship creation
	t.Run("CreateRelationship", func(t *testing.T) {
		rel, err := graph.CreateRelationship(ctx, node1.ID, node2.ID, "KNOWS", map[string]interface{}{
			"since": "2020",
		})
		if err != nil {
			t.Fatalf("Failed to create relationship: %v", err)
		}

		if rel.ID <= 0 {
			t.Error("Expected positive relationship ID")
		}
		if rel.StartNode != node1.ID || rel.EndNode != node2.ID {
			t.Errorf("Expected relationship from %d to %d, got from %d to %d",
				node1.ID, node2.ID, rel.StartNode, rel.EndNode)
		}
		if rel.Type != "KNOWS" {
			t.Errorf("Expected relationship type KNOWS, got %s", rel.Type)
		}
		if rel.Properties["since"] != "2020" {
			t.Errorf("Expected property since=2020, got %v", rel.Properties["since"])
		}
	})

	// Test node finding
	t.Run("FindNodes", func(t *testing.T) {
		nodes, err := graph.FindNodes(ctx, []string{"Person"}, nil)
		if err != nil {
			t.Fatalf("Failed to find nodes: %v", err)
		}

		if len(nodes) < 2 {
			t.Errorf("Expected at least 2 nodes, got %d", len(nodes))
		}

		// Check that we can find specific nodes by properties
		aliceNodes, err := graph.FindNodes(ctx, nil, map[string]interface{}{
			"name": "Alice",
		})
		if err != nil {
			t.Fatalf("Failed to find Alice: %v", err)
		}

		if len(aliceNodes) != 1 {
			t.Errorf("Expected 1 Alice node, got %d", len(aliceNodes))
		}
	})

	// Test relationship finding
	t.Run("FindRelationships", func(t *testing.T) {
		relationships, err := graph.FindRelationships(ctx, "KNOWS", nil)
		if err != nil {
			t.Fatalf("Failed to find relationships: %v", err)
		}

		if len(relationships) < 1 {
			t.Errorf("Expected at least 1 relationship, got %d", len(relationships))
		}

		// Check relationship properties
		rel := relationships[0]
		if rel.Type != "KNOWS" {
			t.Errorf("Expected KNOWS relationship, got %s", rel.Type)
		}
	})

	// Test creating relationship with non-existent nodes (should fail)
	t.Run("CreateRelationshipInvalidNodes", func(t *testing.T) {
		_, err := graph.CreateRelationship(ctx, 999, 1000, "INVALID", nil)
		if err == nil {
			t.Error("Expected error when creating relationship with non-existent nodes")
		}
	})
}
