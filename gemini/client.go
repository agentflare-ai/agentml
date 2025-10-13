package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"maps"

	"google.golang.org/genai"
)

type Client struct {
	genai           *genai.Client
	models          map[ModelName]*Model
	selectionStrategy *ModelSelectionStrategy
}

type ClientOptions = genai.ClientConfig

func NewClient(ctx context.Context, models map[ModelName]*Model, config *genai.ClientConfig) (*Client, error) {
	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, err
	}
	return &Client{
		genai:             client,
		models:            maps.Clone(models),
		selectionStrategy: NewModelSelectionStrategy(),
	}, nil
}

func (c *Client) GenerateContent(ctx context.Context, model ModelName, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	maybeModel, ok := c.models[model]
	if !ok {
		slog.Error("gemini.client.generateContent: model not found", "model", model, "models", c.models)
		return nil, fmt.Errorf("model %s not found", model)
	}
	return maybeModel.generateContent(ctx, c.genai, contents, config)
}

func (c *Client) StreamGenerate(ctx context.Context, model ModelName, contents []*genai.Content, config *genai.GenerateContentConfig, respChan chan<- *genai.GenerateContentResponse) error {
	maybeModel, ok := c.models[model]
	if !ok {
		slog.Error("gemini.client.streamGenerate: model not found", "model", model, "models", c.models)
		return fmt.Errorf("model %s not found", model)
	}
	return maybeModel.generateContentStream(ctx, c.genai, contents, config, respChan)
}

func (c *Client) CountTokens(ctx context.Context, model ModelName, contents []*genai.Content, config *genai.CountTokensConfig) (*genai.CountTokensResponse, error) {
	maybeModel, ok := c.models[model]
	if !ok {
		return nil, fmt.Errorf("model %s not found", model)
	}
	return maybeModel.countTokens(ctx, c.genai, contents, config)
}

func (c *Client) EmbedContent(ctx context.Context, model ModelName, contents []*genai.Content, config *genai.EmbedContentConfig) (*genai.EmbedContentResponse, error) {
	maybeModel, ok := c.models[model]
	if !ok {
		return nil, fmt.Errorf("model %s not found", model)
	}
	return maybeModel.embedContent(ctx, c.genai, contents, config)
}

// GenerateWithAutoSelection automatically selects the best model based on prompt complexity
// and provides fallback to other models if rate limits are encountered.
func (c *Client) GenerateWithAutoSelection(ctx context.Context, prompt string, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, *ModelSelectionResult, error) {
	if c.selectionStrategy == nil {
		return nil, nil, fmt.Errorf("model selection strategy not configured")
	}

	return c.selectionStrategy.GenerateWithFallback(ctx, c, prompt, config)
}

// SetSelectionStrategy allows customization of the model selection strategy.
func (c *Client) SetSelectionStrategy(strategy *ModelSelectionStrategy) {
	c.selectionStrategy = strategy
}
