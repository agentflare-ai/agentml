package ollama

import (
	"context"
	"testing"
)

// Test namespace loader function
func TestNamespaceLoader(t *testing.T) {
	deps := &Deps{}
	loader := Loader(deps)

	if loader == nil {
		t.Error("Loader function returned nil")
	}

	// Test that we can call the loader successfully
	ctx := context.Background()
	ns, err := loader(ctx, nil, nil)
	if err != nil {
		t.Errorf("Expected no error when creating namespace, got: %v", err)
	}
	if ns == nil {
		t.Error("Expected namespace to be created")
	}
}

// Test namespace URI
func TestNamespace_URI(t *testing.T) {
	deps := &Deps{}
	loader := Loader(deps)
	ctx := context.Background()

	ns, err := loader(ctx, nil, nil)
	if err != nil {
		t.Skip("Cannot test namespace URI due to setup error:", err)
		return
	}

	expected := "http://www.gogo-agent.dev/ollama"
	if ns.URI() != expected {
		t.Errorf("Expected URI %s, got %s", expected, ns.URI())
	}
}

// Test namespace unload
func TestNamespace_Unload(t *testing.T) {
	deps := &Deps{}
	loader := Loader(deps)
	ctx := context.Background()

	ns, err := loader(ctx, nil, nil)
	if err != nil {
		t.Skip("Cannot test namespace unload due to setup error:", err)
		return
	}

	err = ns.Unload(ctx)
	if err != nil {
		t.Errorf("Unload should not return error, got: %v", err)
	}
}
