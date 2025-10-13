package ollama

import (
	"context"
	"testing"

	"github.com/agentflare-ai/agentml"
)

// Simple compilation test - verify that the package builds correctly
func TestPackageCompilation(t *testing.T) {
	// This test verifies that all types and functions compile correctly
	// without requiring complex mock implementations

	// Test that Generate implements scxml.Executor
	var _ agentml.Executor = (*Generate)(nil)

	// Test that we can create a basic Generate struct
	generate := &Generate{
		Model:    "llama3.2",
		Prompt:   "test prompt",
		Location: "result",
	}

	if generate.Model != "llama3.2" {
		t.Errorf("Expected model 'llama3.2', got: %s", generate.Model)
	}

	// Test that NewGenerate function exists and can be called with nil (should fail)
	_, err := NewGenerate(context.Background(), nil)
	if err == nil {
		t.Error("Expected error with nil element, got nil")
	}
}

// Test namespace functions exist
func TestNamespaceFunctions(t *testing.T) {
	// Test that Loader function exists
	deps := &Deps{}
	loader := Loader(deps)
	if loader == nil {
		t.Error("Loader function returned nil")
	}

	// Test that we can create a namespace (will fail without proper setup but verifies function exists)
	ctx := context.Background()
	ns, err := loader(ctx, nil, nil)
	if err != nil {
		// Expected to fail without proper interpreter, but function should exist
		t.Logf("Namespace creation failed as expected: %v", err)
	} else if ns == nil {
		t.Error("Namespace creation returned nil")
	}
}

// Test client functions exist
func TestClientFunctions(t *testing.T) {
	// Test that NewClient function exists
	ctx := context.Background()
	models := map[ModelName]*Model{
		Llama3_2: NewModel(Llama3_2, false),
	}

	client, err := NewClient(ctx, models, nil)
	if err != nil {
		t.Logf("Client creation failed (expected due to no real API): %v", err)
	} else if client == nil {
		t.Error("Client creation returned nil")
	}

	// Test that model creation works
	model := NewModel(Llama3_2, true)
	if model.Name != "llama3.2" {
		t.Errorf("Expected model name 'llama3.2', got: %s", model.Name)
	}
	if !model.Stream {
		t.Error("Expected streaming to be enabled")
	}
}

// Test that all model constants are defined
func TestModelConstants(t *testing.T) {
	models := []ModelName{
		Llama3_2,
		Llama3_1,
		Codellama,
		Mistral,
	}

	for _, model := range models {
		if string(model) == "" {
			t.Error("Model constant is empty")
		}
	}
}

// Test template processing
func TestGenerate_processTemplate(t *testing.T) {
	generate := &Generate{}

	t.Run("PlainText", func(t *testing.T) {
		result, err := generate.processTemplate("Hello, World!", nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got: %s", result)
		}
	})

	t.Run("SimpleTemplate", func(t *testing.T) {
		data := map[string]any{"name": "Alice"}
		result, err := generate.processTemplate("Hello, {{.name}}!", data)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "Hello, Alice!" {
			t.Errorf("Expected 'Hello, Alice!', got: %s", result)
		}
	})

	t.Run("NoTemplate", func(t *testing.T) {
		result, err := generate.processTemplate("No template here", nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "No template here" {
			t.Errorf("Expected unchanged text, got: %s", result)
		}
	})
}
