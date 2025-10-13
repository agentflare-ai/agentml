# Makefile for knowledge package with SQLite extensions support

# Build tags to enable SQLite extension loading
SQLITE_TAGS = -tags "sqlite_allow_extension_load"

# CGO flags to enable extension loading
CGO_FLAGS = CGO_CFLAGS="-DSQLITE_ENABLE_LOAD_EXTENSION=1"

# Extension paths
GRAPH_EXT_PATH = extensions/sqlite_graph
VEC_EXT_PATH = extensions/sqlite_vec

# Default target
.PHONY: all
all: generate

# Generate embedded assets
.PHONY: generate
generate: build-graph-extension build-vec-extension
	@echo "Running go generate..."
	@go generate ./...

# Build graph extension
.PHONY: build-graph-extension
build-graph-extension:
	@echo "Building SQLite graph extension..."
	@cd $(GRAPH_EXT_PATH) && make clean && make -C _deps && make -C src
	@echo "Copying graph extension to knowledge package..."
	@if [ "$(shell uname -s)" = "Darwin" ]; then \
		cp $(GRAPH_EXT_PATH)/build/libgraph.dylib graph_extension.so; \
	else \
		cp $(GRAPH_EXT_PATH)/build/libgraph.so graph_extension.so; \
	fi

# Build vec extension  
.PHONY: build-vec-extension
build-vec-extension:
	@echo "Building SQLite vec extension..."
	@cd $(VEC_EXT_PATH) && make clean && make loadable
	@echo "Copying vec extension to knowledge package..."
	@if [ "$(shell uname -s)" = "Darwin" ]; then \
		cp $(VEC_EXT_PATH)/dist/vec0.dylib vec_extension.so; \
	else \
		cp $(VEC_EXT_PATH)/dist/vec0.so vec_extension.so; \
	fi

# Build and generate
.PHONY: build
build: generate
	$(CGO_FLAGS) go build $(SQLITE_TAGS) ./...

# Test with extension loading enabled
.PHONY: test
test: generate
	go test -v ./...

# Test with coverage
.PHONY: test-coverage
test-coverage: generate
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run specific graph tests
.PHONY: test-graph
test-graph: generate
	go test -v -run TestGraph ./...

# Run specific vector tests
.PHONY: test-vector
test-vector: generate
	go test -v -run TestVec ./...

# Clean generated files
.PHONY: clean
clean:
	@rm -f graph_extension.so
	@rm -f vec_extension.so
	@rm -f coverage.out coverage.html
	@rm -f *.db
