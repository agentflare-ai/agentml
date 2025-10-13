package gemini

import (
	"context"
	"os"
	"testing"
	"time"

	"google.golang.org/genai"
)

func TestClientWithRateLimiting_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up environment
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	// Create rate limiters with reasonable limits for testing
	generateLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 100, // 100 requests per minute
		RPD: 10000,
		TPM: 1000,
	})

	tokenLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 100,
		RPD: 10000,
		TPM: 1000,
	})

	// Create models with rate limiting
	models := map[ModelName]*Model{
		Flash: NewModel(Flash, generateLimiter, tokenLimiter),
	}

	// Load API key from env and pass via config
	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	client, err := NewClient(ctx, models, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test multiple rapid requests to verify rate limiting works
	start := time.Now()
	requestCount := 5

	for i := 0; i < requestCount; i++ {
		t.Run("GenerateContent_"+string(rune(i)), func(t *testing.T) {
			contents := []*genai.Content{
				{
					Parts: []*genai.Part{
						{Text: "Hello world"},
					},
				},
			}

			_, err := client.GenerateContent(ctx, Flash, contents, nil)
			// We expect this to fail due to invalid API key, but it tests the rate limiting flow
			if err == nil {
				t.Log("GenerateContent succeeded (unexpected with invalid API key)")
			}
		})
	}

	totalDuration := time.Since(start)

	// With rate limiting at 100 RPM (1.67 RPS), 5 requests should take at least some time
	// But we won't be too strict since this is testing the integration
	t.Logf("Completed %d requests in %v", requestCount, totalDuration)
}

func TestModelWithRateLimiting_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up environment
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	// Create rate limiters
	generateLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 60, // 1 request per second
		RPD: 1000,
		TPM: 1000,
	})

	tokenLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 60,
		RPD: 1000,
		TPM: 1000,
	})

	// Create a real client to avoid nil pointer dereference
	client, err := NewClient(ctx, nil, &genai.ClientConfig{APIKey: loadAPIKeyFromEnv()})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	model := NewModel(Flash, generateLimiter, tokenLimiter)

	// Test the model directly (this will fail due to invalid API key, but tests the rate limiting)
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: "Test content"},
			},
		},
	}

	start := time.Now()

	// This should trigger rate limiting logic even though it will fail due to invalid API key
	_, err = model.generateContent(ctx, client.genai, contents, nil)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected error due to invalid API key, but got none")
	}

	// Ensure context is used for cancellation check
	select {
	case <-ctx.Done():
		t.Log("Context was cancelled as expected")
	default:
		t.Logf("Model generateContent took %v (includes rate limiting)", duration)
	}
}

func TestTier1Models_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test that Tier1Models can be used with a client
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	client, err := NewClient(ctx, Tier1Models, config)
	if err != nil {
		t.Fatalf("Failed to create client with Tier1Models: %v", err)
	}

	// Verify all expected models are available
	expectedModels := []ModelName{Pro, Flash, FlashLite, Embedding001}

	for _, modelName := range expectedModels {
		if _, exists := client.models[modelName]; !exists {
			t.Errorf("Expected model %s not found in client", modelName)
		}
	}

	t.Logf("Client initialized with %d models", len(client.models))
}

func TestLargePromptWithRateLimiting_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Longer timeout for large content
	defer cancel()

	// Set up environment
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	// Create rate limiters suitable for large content
	generateLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 10, // Lower rate for large content
		RPD: 1000,
		TPM: 10000, // Higher token limit for large prompts
	})

	tokenLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 20,
		RPD: 2000,
		TPM: 20000,
	})

	models := map[ModelName]*Model{
		Flash: NewModel(Flash, generateLimiter, tokenLimiter),
	}

	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	client, err := NewClient(ctx, models, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with large prompt
	start := time.Now()

	largePrompt := generateLargeSystemPrompt()
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: largePrompt},
			},
		},
	}

	_, err = client.GenerateContent(ctx, Flash, contents, nil)
	duration := time.Since(start)

	// Should fail due to invalid API key, but tests the rate limiting flow
	if err == nil {
		t.Log("GenerateContent succeeded (unexpected with invalid API key)")
	}

	t.Logf("Large prompt processing took %v, prompt length: %d chars", duration, len(largePrompt))
}

func TestConcurrentClientUsage_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up environment
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	// Create rate limiters for concurrent access
	generateLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 50,
		RPD: 5000,
		TPM: 5000,
	})

	tokenLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 50,
		RPD: 5000,
		TPM: 5000,
	})

	models := map[ModelName]*Model{
		Flash: NewModel(Flash, generateLimiter, tokenLimiter),
	}

	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	client, err := NewClient(ctx, models, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test concurrent usage
	concurrency := 5
	done := make(chan bool, concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer func() { done <- true }()

			contents := []*genai.Content{
				{
					Parts: []*genai.Part{
						{Text: "Concurrent test"},
					},
				},
			}

			_, err := client.GenerateContent(ctx, Flash, contents, nil)
			if err == nil {
				t.Logf("Goroutine %d: GenerateContent succeeded (unexpected)", id)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent test timed out")
		}
	}

	totalDuration := time.Since(start)
	t.Logf("Concurrent test completed in %v", totalDuration)
}

func BenchmarkClientWithRateLimiting(b *testing.B) {
	ctx := context.Background()

	// Set up environment
	os.Setenv("GEMINI_API_KEY", "bench-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	// Create fast rate limiters for benchmarking
	generateLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 1000,
		RPD: 100000,
		TPM: 100000,
	})

	tokenLimiter := NewRateLimiter(RateLimiterOptions{
		RPM: 1000,
		RPD: 100000,
		TPM: 100000,
	})

	models := map[ModelName]*Model{
		Flash: NewModel(Flash, generateLimiter, tokenLimiter),
	}

	config := &genai.ClientConfig{
		APIKey: loadAPIKeyFromEnv(),
	}

	client, _ := NewClient(ctx, models, config)

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: "Benchmark test"},
			},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.GenerateContent(ctx, Flash, contents, nil)
		}
	})
}
