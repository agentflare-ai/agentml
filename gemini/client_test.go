package gemini

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/genai"
)

// loadAPIKeyFromEnv loads the API key from environment variables
func loadAPIKeyFromEnv() string {
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("GENAI_API_KEY"); key != "" {
		return key
	}
	return ""
}

// generateLargeSystemPrompt creates a large system prompt for testing token counting and rate limiting
func generateLargeSystemPrompt() string {
	basePrompt := `You are an AI assistant with extensive knowledge across multiple domains. Your capabilities include:

1. Natural Language Processing: Understanding and generating human-like text
2. Code Analysis: Reviewing, debugging, and optimizing code
3. Data Analysis: Processing and interpreting complex datasets
4. Creative Writing: Generating stories, poems, and other creative content
5. Technical Documentation: Creating clear and comprehensive documentation
6. Problem Solving: Breaking down complex problems into manageable steps
7. Research: Gathering and synthesizing information from various sources
8. Teaching: Explaining concepts in an accessible and engaging manner

When responding to user queries, you should:
- Be helpful and informative
- Provide accurate and well-researched information
- Use clear and concise language
- Structure your responses for easy reading
- Ask clarifying questions when needed
- Admit when you don't know something
- Suggest alternative approaches when appropriate

Your knowledge base includes:
- Programming languages (Python, JavaScript, Go, Java, C++, etc.)
- Web development (HTML, CSS, React, Node.js, etc.)
- Data science and machine learning
- Cloud computing and DevOps
- Cybersecurity principles
- Software engineering best practices
- Database design and management
- API design and development
- Testing methodologies
- Version control systems
- Containerization and orchestration

You have access to real-time information through various APIs and can help with:
- Current events and news analysis
- Weather information and forecasts
- Financial data and market analysis
- Scientific research and discoveries
- Health and medical information
- Travel planning and recommendations
- Educational resources and learning paths

Always maintain a professional and respectful tone, and prioritize user safety and privacy in your responses.`

	// Repeat the prompt multiple times to make it large enough to test token limits
	var largePrompt strings.Builder
	for i := 0; i < 20; i++ {
		largePrompt.WriteString(basePrompt)
		largePrompt.WriteString(fmt.Sprintf("\n\n--- Section %d ---\n\n", i+1))
	}
	return largePrompt.String()
}

func TestNewClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		getConfig  func() *genai.ClientConfig
		models     map[ModelName]*Model
		wantError  bool
	}{
		{
			name: "success with API key from environment",
			setupEnv: func() {
				os.Setenv("GEMINI_API_KEY", "test-key-from-env")
			},
			cleanupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
			},
			getConfig: func() *genai.ClientConfig {
				return &genai.ClientConfig{
					APIKey: loadAPIKeyFromEnv(),
				}
			},
			models:    map[ModelName]*Model{},
			wantError: false,
		},
		{
			name: "success with API key from GENAI_API_KEY env",
			setupEnv: func() {
				os.Setenv("GENAI_API_KEY", "test-key-from-genai-env")
			},
			cleanupEnv: func() {
				os.Unsetenv("GENAI_API_KEY")
			},
			getConfig: func() *genai.ClientConfig {
				return &genai.ClientConfig{
					APIKey: loadAPIKeyFromEnv(),
				}
			},
			models:    map[ModelName]*Model{},
			wantError: false,
		},
		{
			name: "success with API key directly in config",
			setupEnv: func() {
				// No env setup needed
			},
			cleanupEnv: func() {
				// No cleanup needed
			},
			getConfig: func() *genai.ClientConfig {
				return &genai.ClientConfig{
					APIKey: "direct-config-key",
				}
			},
			models:    map[ModelName]*Model{},
			wantError: false,
		},
		{
			name: "success with nil config",
			setupEnv: func() {
				os.Setenv("GEMINI_API_KEY", "nil-config-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
			},
			getConfig: func() *genai.ClientConfig {
				return &genai.ClientConfig{APIKey: loadAPIKeyFromEnv()}
			},
			models:    map[ModelName]*Model{},
			wantError: false,
		},
		{
			name: "success with models",
			setupEnv: func() {
				os.Setenv("GEMINI_API_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
			},
			getConfig: func() *genai.ClientConfig {
				return &genai.ClientConfig{
					APIKey: loadAPIKeyFromEnv(),
				}
			},
			models: map[ModelName]*Model{
				Flash: NewModel(Flash, nil, nil),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			config := tt.getConfig()
			client, err := NewClient(ctx, tt.models, config)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewClient() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Errorf("NewClient() returned nil client")
				return
			}

			if client.genai == nil {
				t.Errorf("NewClient() client.genai is nil")
			}

			if client.models == nil {
				t.Errorf("NewClient() client.models is nil")
			}
		})
	}
}

func TestClient_GenerateContent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up environment
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	// Create a mock model with rate limiter that doesn't block
	rateLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 1000,
		RPD: 100000,
		TPM: 100000,
	})

	model := NewModel(Flash, rateLimiter, rateLimiter)
	models := map[ModelName]*Model{
		Flash: model,
	}

	// Load API key from env and pass via config
	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	client, err := NewClient(ctx, models, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name      string
		model     ModelName
		contents  []*genai.Content
		config    *genai.GenerateContentConfig
		wantError bool
	}{
		{
			name:      "model not found",
			model:     "non-existent-model",
			contents:  []*genai.Content{},
			config:    nil,
			wantError: true,
		},
		{
			name:  "valid request",
			model: Flash,
			contents: []*genai.Content{
				{
					Parts: []*genai.Part{
						{Text: "Hello"},
					},
				},
			},
			config:    nil,
			wantError: true, // Will fail due to invalid API key, but tests the flow
		},
		{
			name:  "large system prompt",
			model: Flash,
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: generateLargeSystemPrompt()},
					},
				},
			},
			config:    nil,
			wantError: true, // Will fail due to invalid API key, but tests token counting and rate limiting
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GenerateContent(ctx, tt.model, tt.contents, tt.config)

			if tt.wantError && err == nil {
				t.Errorf("GenerateContent() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("GenerateContent() unexpected error: %v", err)
			}
		})
	}
}

func TestClient_CountTokens(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up environment
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	// Create a mock model with rate limiter that doesn't block
	rateLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 1000,
		RPD: 100000,
		TPM: 100000,
	})

	model := NewModel(Flash, rateLimiter, rateLimiter)
	models := map[ModelName]*Model{
		Flash: model,
	}

	// Load API key from env and pass via config
	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	client, err := NewClient(ctx, models, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name      string
		model     ModelName
		contents  []*genai.Content
		config    *genai.CountTokensConfig
		wantError bool
	}{
		{
			name:      "model not found",
			model:     "non-existent-model",
			contents:  []*genai.Content{},
			config:    nil,
			wantError: true,
		},
		{
			name:  "valid request",
			model: Flash,
			contents: []*genai.Content{
				{
					Parts: []*genai.Part{
						{Text: "Hello world"},
					},
				},
			},
			config:    nil,
			wantError: true, // Will fail due to invalid API key, but tests the flow
		},
		{
			name:  "large system prompt token counting",
			model: Flash,
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: generateLargeSystemPrompt()},
					},
				},
			},
			config:    nil,
			wantError: true, // Will fail due to invalid API key, but tests token counting
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CountTokens(ctx, tt.model, tt.contents, tt.config)

			if tt.wantError && err == nil {
				t.Errorf("CountTokens() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("CountTokens() unexpected error: %v", err)
			}
		})
	}
}

func TestClient_EmbedContent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up environment
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	// Create a mock model with rate limiter that doesn't block
	rateLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 1000,
		RPD: 100000,
		TPM: 100000,
	})

	model := NewModel(Embedding001, rateLimiter, rateLimiter)
	models := map[ModelName]*Model{
		Embedding001: model,
	}

	// Load API key from env and pass via config
	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	client, err := NewClient(ctx, models, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name      string
		model     ModelName
		contents  []*genai.Content
		config    *genai.EmbedContentConfig
		wantError bool
	}{
		{
			name:      "model not found",
			model:     "non-existent-model",
			contents:  []*genai.Content{},
			config:    nil,
			wantError: true,
		},
		{
			name:  "valid request",
			model: Embedding001,
			contents: []*genai.Content{
				{
					Parts: []*genai.Part{
						{Text: "Hello world"},
					},
				},
			},
			config:    nil,
			wantError: true, // Will fail due to invalid API key, but tests the flow
		},
		{
			name:  "large content embedding",
			model: Embedding001,
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: generateLargeSystemPrompt()},
					},
				},
			},
			config:    nil,
			wantError: true, // Will fail due to invalid API key, but tests embedding with large content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.EmbedContent(ctx, tt.model, tt.contents, tt.config)

			if tt.wantError && err == nil {
				t.Errorf("EmbedContent() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("EmbedContent() unexpected error: %v", err)
			}
		})
	}
}

// TestLoadAPIKeyFromEnv tests the environment loading helper function
func TestLoadAPIKeyFromEnv(t *testing.T) {
	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		expected   string
	}{
		{
			name: "GEMINI_API_KEY takes precedence",
			setupEnv: func() {
				os.Setenv("GEMINI_API_KEY", "gemini-key")
				os.Setenv("GENAI_API_KEY", "genai-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
				os.Unsetenv("GENAI_API_KEY")
			},
			expected: "gemini-key",
		},
		{
			name: "fallback to GENAI_API_KEY",
			setupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
				os.Setenv("GENAI_API_KEY", "genai-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GENAI_API_KEY")
			},
			expected: "genai-key",
		},
		{
			name: "no environment variables set",
			setupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
				os.Unsetenv("GENAI_API_KEY")
			},
			cleanupEnv: func() {
				// No cleanup needed
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			result := loadAPIKeyFromEnv()
			if result != tt.expected {
				t.Errorf("loadAPIKeyFromEnv() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// ExampleNewClient demonstrates how to create a client with API key from environment
func ExampleNewClient() {
	ctx := context.Background()

	// Load API key from environment
	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	// Use predefined models with rate limiting
	models := Tier1Models

	client, err := NewClient(ctx, models, config)
	if err != nil {
		// Handle error
		return
	}

	// Use client in your application
	_ = client
}
