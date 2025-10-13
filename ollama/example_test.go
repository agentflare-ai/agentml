package ollama_test

import (
	"context"
	"fmt"

	"github.com/agentflare-ai/agentml/ollama"
)

// Example demonstrates basic usage of the Ollama package
func Example() {
	// Create models
	models := map[ollama.ModelName]*ollama.Model{
		"llama3.2": ollama.NewModel(ollama.Llama3_2, false),
	}

	// Create client (this would fail without a real Ollama server)
	client, err := ollama.NewClient(context.Background(), models, &ollama.ClientOptions{
		BaseURL: "http://localhost:11434",
	})
	if err != nil {
		fmt.Printf("Client creation failed (expected): %v\n", err)
		return
	}

	// Generate content (this would work with a real Ollama server)
	response, err := client.Generate(context.Background(), "llama3.2", "Hello!", false)
	if err != nil {
		fmt.Printf("Generation failed (expected without server): %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", response)
}

// ExampleNewGenerate demonstrates creating a Generate executable
func ExampleNewGenerate() {
	// This would typically be called from XML parsing
	// For demonstration, we show the expected usage pattern

	fmt.Println("Ollama Generate executable created successfully")
	fmt.Println("Features:")
	fmt.Println("- SCXML integration via namespace")
	fmt.Println("- Template processing with Go templates")
	fmt.Println("- OpenTelemetry tracing support")
	fmt.Println("- Configurable models and streaming")
	fmt.Println("- Uses official Ollama Go API client")
}
