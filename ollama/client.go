package ollama

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/ollama/ollama/api"
)

type Client struct {
	apiClient *api.Client
	models    map[ModelName]*Model
}

type ClientOptions struct {
	BaseURL string
}

func NewClient(ctx context.Context, models map[ModelName]*Model, options *ClientOptions) (*Client, error) {
	if options == nil {
		options = &ClientOptions{}
	}

	apiClient, err := NewOllamaClient(options.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	// Initialize models map if nil
	if models == nil {
		models = make(map[ModelName]*Model)
	}

	return &Client{
		apiClient: apiClient,
		models:    maps.Clone(models),
	}, nil
}

func (c *Client) Generate(ctx context.Context, model ModelName, prompt string, stream bool) (string, error) {
	maybeModel, ok := c.models[model]
	if !ok {
		// Dynamically register the model if not found
		slog.Debug("ollama.client.generate: dynamically registering model", "model", model)
		maybeModel = &Model{
			Name:   string(model),
			Stream: stream,
		}
		c.models[model] = maybeModel
	}

	req := &api.GenerateRequest{
		Model:  maybeModel.Name,
		Prompt: prompt,
		Stream: &stream,
	}

	var fullResponse strings.Builder
	err := c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fullResponse.WriteString(resp.Response)
		return nil
	})

	if err != nil {
		return "", err
	}

	return fullResponse.String(), nil
}

func (c *Client) Chat(ctx context.Context, model ModelName, messages []api.Message, stream bool) (*api.ChatResponse, error) {
	maybeModel, ok := c.models[model]
	if !ok {
		// Dynamically register the model if not found
		slog.Debug("ollama.client.chat: dynamically registering model", "model", model)
		maybeModel = &Model{
			Name:   string(model),
			Stream: stream,
		}
		c.models[model] = maybeModel
	}

	req := &api.ChatRequest{
		Model:    maybeModel.Name,
		Messages: messages,
		Stream:   &stream,
	}

	var finalResponse *api.ChatResponse
	err := c.apiClient.Chat(ctx, req, func(resp api.ChatResponse) error {
		if resp.Done {
			finalResponse = &resp
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if finalResponse == nil {
		return nil, fmt.Errorf("no complete response received")
	}

	return finalResponse, nil
}

func (c *Client) ChatWithTools(ctx context.Context, model ModelName, messages []api.Message, tools []api.Tool, stream bool) (*api.ChatResponse, error) {
	maybeModel, ok := c.models[model]
	if !ok {
		// Dynamically register the model if not found
		slog.Debug("ollama.client.chatWithTools: dynamically registering model", "model", model)
		maybeModel = &Model{
			Name:   string(model),
			Stream: stream,
		}
		c.models[model] = maybeModel
	}

	req := &api.ChatRequest{
		Model:    maybeModel.Name,
		Messages: messages,
		Tools:    tools,
		Stream:   &stream,
	}

	var finalResponse *api.ChatResponse
	err := c.apiClient.Chat(ctx, req, func(resp api.ChatResponse) error {
		if resp.Done {
			finalResponse = &resp
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if finalResponse == nil {
		return nil, fmt.Errorf("no complete response received")
	}

	return finalResponse, nil
}

// Deps holds dependencies for Ollama executables
type Deps struct {
	Client *Client
}
