package ollama

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

// Model represents an Ollama model configuration
type Model struct {
	Name   string
	Stream bool
}

// ModelName represents an Ollama model name
type ModelName string

const (
	// Common Ollama models
	Llama3_2  ModelName = "llama3.2"
	Llama3_1  ModelName = "llama3.1"
	Codellama ModelName = "codellama"
	Mistral   ModelName = "mistral"
)

// NewModel creates a new Ollama model configuration
func NewModel(name ModelName, stream bool) *Model {
	return &Model{
		Name:   string(name),
		Stream: stream,
	}
}

// NewOllamaClient creates a new Ollama client using the official API
func NewOllamaClient(baseURL string) (*api.Client, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := api.NewClient(parsedURL, http.DefaultClient)
	return client, nil
}
