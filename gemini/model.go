package gemini

import (
	"context"

	"google.golang.org/genai"
)

type Model struct {
	Name                  ModelName
	GenerateRateLimiter   *RateLimiter
	TokenCountRateLimiter *RateLimiter
}

type ModelName string

const (
	FlashLite    ModelName = "gemini-2.5-flash-lite"
	Pro          ModelName = "gemini-2.5-pro"
	Flash        ModelName = "gemini-2.5-flash"
	Ultra        ModelName = "gemini-2.5-pro" // compatibility alias used by coder package
	Embedding001 ModelName = "gemini-embedding-001"
)

func NewModel(name ModelName, generateRateLimit *RateLimiter, tokenCountRateLimit *RateLimiter) *Model {
	return &Model{
		Name:                  name,
		GenerateRateLimiter:   generateRateLimit,
		TokenCountRateLimiter: tokenCountRateLimit,
	}
}

func (m *Model) countTokens(ctx context.Context, client *genai.Client, contents []*genai.Content, config *genai.CountTokensConfig) (*genai.CountTokensResponse, error) {
	if err := m.TokenCountRateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	count, err := client.Models.CountTokens(ctx, string(m.Name), contents, config)
	if err != nil {
		return nil, err
	}

	return count, nil
}

func (m *Model) generateContent(ctx context.Context, client *genai.Client, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	// Count tokens first - provides natural backpressure
	tokenCount, err := m.countTokens(ctx, client, contents, nil)
	if err != nil {
		return nil, err
	}

	// Apply rate limiting after we know token count
	if err := m.GenerateRateLimiter.Wait(ctx, int(tokenCount.TotalTokens)); err != nil {
		return nil, err
	}

	return client.Models.GenerateContent(ctx, string(m.Name), contents, config)
}

func (m *Model) generateContentStream(ctx context.Context, client *genai.Client, contents []*genai.Content, config *genai.GenerateContentConfig, respChan chan<- *genai.GenerateContentResponse) error {
	// Count tokens first - provides natural backpressure
	tokenCount, err := m.countTokens(ctx, client, contents, nil)
	if err != nil {
		return err
	}

	// Apply rate limiting after we know token count
	if err := m.GenerateRateLimiter.Wait(ctx, int(tokenCount.TotalTokens)); err != nil {
		return err
	}

	// Start streaming generation using the new iterator API
	stream := client.Models.GenerateContentStream(ctx, string(m.Name), contents, config)

	// Process streaming responses using the iterator
	for response, err := range stream {
		if err != nil {
			return err
		}

		select {
		case respChan <- response:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (m *Model) embedContent(ctx context.Context, client *genai.Client, contents []*genai.Content, config *genai.EmbedContentConfig) (*genai.EmbedContentResponse, error) {
	tokenCount, err := m.countTokens(ctx, client, contents, nil)
	if err != nil {
		return nil, err
	}
	if err := m.GenerateRateLimiter.Wait(ctx, int(tokenCount.TotalTokens)); err != nil {
		return nil, err
	}
	return client.Models.EmbedContent(ctx, string(m.Name), contents, config)
}

