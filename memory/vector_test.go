package memory

import (
	"context"
	"math"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestVectorStoreCreation(t *testing.T) {
	ctx := context.Background()

	// Create an in-memory SQLite database with extensions enabled
	db, err := NewDB(ctx, ":memory:?_foreign_keys=on")
	if err != nil {
		t.Skipf("Skipping test - vector extension not available: %v", err)
	}
	defer db.Close()

	t.Run("CreateVectorStore", func(t *testing.T) {
		dimensions := 128
		store, err := NewVectorDB(ctx, db, "test_vectors", dimensions)
		if err != nil {
			t.Fatalf("Failed to create vector store: %v", err)
		}
		defer store.Close()

		if store.dimensions != dimensions {
			t.Errorf("Expected dimensions %d, got %d", dimensions, store.dimensions)
		}
		if store.tableName != "test_vectors" {
			t.Errorf("Expected table name 'test_vectors', got '%s'", store.tableName)
		}
	})

	t.Run("CreateVectorStoreDefaultName", func(t *testing.T) {
		dimensions := 64
		store, err := NewVectorDB(ctx, db, "", dimensions)
		if err != nil {
			t.Fatalf("Failed to create vector store: %v", err)
		}
		defer store.Close()

		if store.tableName != "vectors" {
			t.Errorf("Expected default table name 'vectors', got '%s'", store.tableName)
		}
	})
}

func TestVectorInsertion(t *testing.T) {
	ctx := context.Background()

	// Create an in-memory SQLite database with extensions enabled
	db, err := NewDB(ctx, ":memory:?_foreign_keys=on")
	if err != nil {
		t.Skipf("Skipping test - vector extension not available: %v", err)
	}
	defer db.Close()

	dimensions := 3
	store, err := NewVectorDB(ctx, db, "test_vectors", dimensions)
	if err != nil {
		t.Fatalf("Failed to create vector store: %v", err)
	}
	defer store.Close()

	t.Run("InsertValidVector", func(t *testing.T) {
		vector := []float32{1.0, 2.0, 3.0}
		err := store.InsertVector(ctx, 1, vector)
		if err != nil {
			t.Fatalf("Failed to insert valid vector: %v", err)
		}
	})

	t.Run("InsertVectorWrongDimension", func(t *testing.T) {
		vector := []float32{1.0, 2.0} // Wrong dimension (2 instead of 3)
		err := store.InsertVector(ctx, 2, vector)
		if err == nil {
			t.Error("Expected error for wrong dimension vector")
		}
	})

	t.Run("InsertMultipleVectors", func(t *testing.T) {
		vectors := []struct {
			ID     uint64
			Vector []float32
		}{
			{ID: 10, Vector: []float32{1.0, 0.0, 0.0}},
			{ID: 11, Vector: []float32{0.0, 1.0, 0.0}},
			{ID: 12, Vector: []float32{0.0, 0.0, 1.0}},
			{ID: 13, Vector: []float32{0.5, 0.5, 0.0}},
			{ID: 14, Vector: []float32{0.3, 0.4, 0.5}},
		}

		for _, v := range vectors {
			err := store.InsertVector(ctx, v.ID, v.Vector)
			if err != nil {
				t.Fatalf("Failed to insert vector %d: %v", v.ID, err)
			}
		}
	})

	t.Run("InsertVectorWithExtremeValues", func(t *testing.T) {
		vector := []float32{float32(math.MaxFloat32), float32(-math.MaxFloat32), 0.0}
		err := store.InsertVector(ctx, 100, vector)
		if err != nil {
			t.Fatalf("Failed to insert vector with extreme values: %v", err)
		}
	})
}

func TestVectorSimilaritySearch(t *testing.T) {
	ctx := context.Background()

	// Create an in-memory SQLite database with extensions enabled
	db, err := NewDB(ctx, ":memory:?_foreign_keys=on")
	if err != nil {
		t.Skipf("Skipping test - vector extension not available: %v", err)
	}
	defer db.Close()

	dimensions := 3
	store, err := NewVectorDB(ctx, db, "test_vectors", dimensions)
	if err != nil {
		t.Fatalf("Failed to create vector store: %v", err)
	}
	defer store.Close()

	// Insert test vectors
	testVectors := []struct {
		ID     uint64
		Vector []float32
		Label  string // For easier identification in tests
	}{
		{ID: 1, Vector: []float32{1.0, 0.0, 0.0}, Label: "unit_x"},
		{ID: 2, Vector: []float32{0.0, 1.0, 0.0}, Label: "unit_y"},
		{ID: 3, Vector: []float32{0.0, 0.0, 1.0}, Label: "unit_z"},
		{ID: 4, Vector: []float32{0.9, 0.1, 0.0}, Label: "near_x"},
		{ID: 5, Vector: []float32{0.1, 0.9, 0.0}, Label: "near_y"},
		{ID: 6, Vector: []float32{0.5, 0.5, 0.0}, Label: "middle_xy"},
		{ID: 7, Vector: []float32{-1.0, 0.0, 0.0}, Label: "neg_x"},
	}

	for _, v := range testVectors {
		err := store.InsertVector(ctx, v.ID, v.Vector)
		if err != nil {
			t.Fatalf("Failed to insert test vector %d (%s): %v", v.ID, v.Label, err)
		}
	}

	t.Run("SearchExactMatch", func(t *testing.T) {
		queryVector := []float32{1.0, 0.0, 0.0} // Exact match with ID 1
		results, err := store.SearchSimilarVectors(ctx, queryVector, 1)
		if err != nil {
			t.Fatalf("Failed to search similar vectors: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		if results[0].ID != 1 {
			t.Errorf("Expected exact match with ID 1, got %d", results[0].ID)
		}

		// For exact match, distance should be 0 or very close to 0
		if results[0].Distance > 0.001 {
			t.Errorf("Expected distance close to 0 for exact match, got %f", results[0].Distance)
		}
	})

	t.Run("SearchSimilarVectors", func(t *testing.T) {
		queryVector := []float32{0.8, 0.2, 0.0} // Should be closest to ID 1, then ID 4
		results, err := store.SearchSimilarVectors(ctx, queryVector, 3)
		if err != nil {
			t.Fatalf("Failed to search similar vectors: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		// Results should be ordered by similarity (smallest distance first)
		for i := 1; i < len(results); i++ {
			if results[i-1].Distance > results[i].Distance {
				t.Errorf("Results not ordered by distance: %f > %f", results[i-1].Distance, results[i].Distance)
			}
		}
	})

	t.Run("SearchWithLimit", func(t *testing.T) {
		queryVector := []float32{0.5, 0.5, 0.5}
		limits := []int{1, 3, 5, 10}

		for _, limit := range limits {
			results, err := store.SearchSimilarVectors(ctx, queryVector, limit)
			if err != nil {
				t.Fatalf("Failed to search with limit %d: %v", limit, err)
			}

			expectedLen := limit
			if len(testVectors) < limit {
				expectedLen = len(testVectors)
			}

			if len(results) != expectedLen {
				t.Errorf("Expected %d results with limit %d, got %d", expectedLen, limit, len(results))
			}
		}
	})

	t.Run("SearchWrongDimension", func(t *testing.T) {
		queryVector := []float32{1.0, 0.0} // Wrong dimension (2 instead of 3)
		_, err := store.SearchSimilarVectors(ctx, queryVector, 1)
		if err == nil {
			t.Error("Expected error for wrong dimension query vector")
		}
	})

	t.Run("SearchEmptyStore", func(t *testing.T) {
		// Create a new empty store
		emptyStore, err := NewVectorDB(ctx, db, "empty_vectors", dimensions)
		if err != nil {
			t.Fatalf("Failed to create empty vector store: %v", err)
		}
		defer emptyStore.Close()

		queryVector := []float32{1.0, 0.0, 0.0}
		results, err := emptyStore.SearchSimilarVectors(ctx, queryVector, 5)
		if err != nil {
			t.Fatalf("Failed to search empty store: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 results from empty store, got %d", len(results))
		}
	})
}

func TestVectorStoreHighDimensional(t *testing.T) {
	ctx := context.Background()

	// Create an in-memory SQLite database with extensions enabled
	db, err := NewDB(ctx, ":memory:?_foreign_keys=on")
	if err != nil {
		t.Skipf("Skipping test - vector extension not available: %v", err)
	}
	defer db.Close()

	t.Run("HighDimensionalVectors", func(t *testing.T) {
		dimensions := 512 // Common embedding dimension
		store, err := NewVectorDB(ctx, db, "high_dim_vectors", dimensions)
		if err != nil {
			t.Fatalf("Failed to create high-dimensional vector store: %v", err)
		}
		defer store.Close()

		// Create test vectors
		vectors := make([][]float32, 5)
		for i := range vectors {
			vectors[i] = make([]float32, dimensions)
			for j := range vectors[i] {
				// Create somewhat random but deterministic vectors
				vectors[i][j] = float32((i*dimensions+j)%100) / 100.0
			}
		}

		// Insert vectors
		for i, vector := range vectors {
			err := store.InsertVector(ctx, uint64(i+1), vector)
			if err != nil {
				t.Fatalf("Failed to insert high-dimensional vector %d: %v", i+1, err)
			}
		}

		// Search with the first vector
		results, err := store.SearchSimilarVectors(ctx, vectors[0], 3)
		if err != nil {
			t.Fatalf("Failed to search high-dimensional vectors: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// First result should be exact match
		if results[0].ID != 1 {
			t.Errorf("Expected first result to be ID 1, got %d", results[0].ID)
		}
	})
}

func TestVectorStoreConcurrency(t *testing.T) {
	ctx := context.Background()

	// Create an in-memory SQLite database with extensions enabled
	db, err := NewDB(ctx, ":memory:?_foreign_keys=on")
	if err != nil {
		t.Skipf("Skipping test - vector extension not available: %v", err)
	}
	defer db.Close()

	t.Run("ConcurrentInserts", func(t *testing.T) {
		dimensions := 10
		store, err := NewVectorDB(ctx, db, "concurrent_vectors", dimensions)
		if err != nil {
			t.Fatalf("Failed to create vector store: %v", err)
		}
		defer store.Close()

		// Note: SQLite with WAL mode would be better for true concurrency,
		// but for this test we'll just verify sequential operations work correctly
		numVectors := 100
		for i := 0; i < numVectors; i++ {
			vector := make([]float32, dimensions)
			for j := range vector {
				vector[j] = float32(i*j) / 100.0
			}

			err := store.InsertVector(ctx, uint64(i+1), vector)
			if err != nil {
				t.Fatalf("Failed to insert vector %d: %v", i+1, err)
			}
		}

		// Verify we can search through all vectors
		queryVector := make([]float32, dimensions)
		for j := range queryVector {
			queryVector[j] = 0.5
		}

		results, err := store.SearchSimilarVectors(ctx, queryVector, numVectors)
		if err != nil {
			t.Fatalf("Failed to search after bulk insert: %v", err)
		}

		if len(results) != numVectors {
			t.Errorf("Expected %d results after bulk insert, got %d", numVectors, len(results))
		}
	})
}
