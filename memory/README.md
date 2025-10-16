# Memory

A high-performance memory and vector database package for Go, built on SQLite with advanced graph and vector search capabilities.

## Features

* **SQLite-based**: Reliable, embedded database with ACID guarantees
* **Vector Search**: Built-in vector similarity search with multiple distance metrics
* **Graph Database**: Advanced graph operations with Cypher-like query support
* **Schema Management**: Automatic schema creation and migration
* **Embedding Support**: Direct integration for AI/ML embeddings
* **High Performance**: Optimized for both read and write operations

## Extensions

This package includes two powerful SQLite extensions:

### SQLite Graph Extension

* **Cypher Query Support**: Neo4j-compatible graph queries
* **Graph Algorithms**: Shortest path, centrality measures, community detection
* **Performance Optimized**: Efficient graph traversal and storage
* **Full Documentation**: Comprehensive API reference and tutorials

### SQLite Vector Extension

* **Vector Similarity Search**: Cosine, Euclidean, and dot product distances
* **Index Optimization**: Fast approximate nearest neighbor search
* **Batch Operations**: Efficient bulk vector operations
* **Integration Ready**: Works seamlessly with embedding models

## Installation

```bash
go get github.com/agentflare-ai/agentml/memory
```

**Note**: This package requires CGO and SQLite development headers.

## Quick Start

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    "github.com/agentflare-ai/agentml/memory"
)

func main() {
    // Open database
    db, err := sql.Open("sqlite3", "memory.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Load graph extension
    if err := memory.LoadGraphExtension(db); err != nil {
        log.Fatal(err)
    }
    
    // Load vector extension  
    if err := memory.LoadVectorExtension(db); err != nil {
        log.Fatal(err)
    }
    
    // Use the database for graph and vector operations...
}
```

## Graph Operations

```sql
-- Create nodes and relationships
CREATE (a:Person {name: 'Alice', age: 30})
CREATE (b:Person {name: 'Bob', age: 25})
CREATE (a)-[:KNOWS]->(b)

-- Query the graph
MATCH (p:Person)-[:KNOWS]->(friend)
WHERE p.age > 25
RETURN p.name, friend.name
```

## Vector Operations

```sql
-- Create a vector table
CREATE TABLE embeddings (
    id INTEGER PRIMARY KEY,
    content TEXT,
    vector VECTOR(384)
);

-- Insert vectors
INSERT INTO embeddings (content, vector) 
VALUES ('example text', vector('[0.1, 0.2, 0.3, ...]'));

-- Similarity search
SELECT content, vector_distance(vector, '[0.1, 0.2, ...]') as distance
FROM embeddings
ORDER BY distance
LIMIT 10;
```

## Building Extensions

The package includes build tools for compiling the native extensions:

```bash
cd memory
make build-extensions
```

## Performance

* **Graph queries**: Optimized for complex traversals
* **Vector search**: Sub-millisecond similarity search on large datasets
* **Memory efficient**: Minimal overhead over base SQLite
* **Concurrent safe**: Full support for concurrent read/write operations

## Documentation

* [Graph Extension API](extensions/sqlite_graph/docs/API_REFERENCE.md)
* [Performance Tuning Guide](extensions/sqlite_graph/docs/PERFORMANCE_TUNING_GUIDE.md)
* [Migration Guide](extensions/sqlite_graph/docs/MIGRATION_GUIDE.md)
* [Tutorials](extensions/sqlite_graph/docs/TUTORIALS.md)

## License

This project is part of the AgentML ecosystem.
