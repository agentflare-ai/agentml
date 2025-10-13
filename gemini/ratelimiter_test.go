package gemini

import (
	"context"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name    string
		options RateLimiterOptions
		wantRPM rate.Limit
		wantRPD rate.Limit
		wantTPM rate.Limit
	}{
		{
			name: "standard tier 1 limits",
			options: RateLimiterOptions{
				RPM: 150,
				RPD: 2_000_000,
				TPM: 10_000,
			},
			wantRPM: 150.0 / 60.0,          // RPM converted to per-second rate
			wantRPD: 2_000_000.0 / 86400.0, // RPD converted to per-second rate
			wantTPM: 10_000.0 / 60.0,       // TPM converted to per-second rate
		},
		{
			name: "high volume limits",
			options: RateLimiterOptions{
				RPM: 1000,
				RPD: 10_000_000,
				TPM: 100_000,
			},
			wantRPM: 1000.0 / 60.0,
			wantRPD: 10_000_000.0 / 86400.0,
			wantTPM: 100_000.0 / 60.0,
		},
		{
			name: "zero limits",
			options: RateLimiterOptions{
				RPM: 0,
				RPD: 0,
				TPM: 0,
			},
			wantRPM: 0,
			wantRPD: 0,
			wantTPM: 0,
		},
		{
			name: "infinite TPM",
			options: RateLimiterOptions{
				RPM: 100,
				RPD: 1000,
				TPM: rate.Inf,
			},
			wantRPM: 100.0 / 60.0,
			wantRPD: 1000.0 / 86400.0,
			wantTPM: rate.Inf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.options)

			if limiter == nil {
				t.Fatal("NewRateLimiter() returned nil")
			}

			if limiter.rpm == nil {
				t.Error("rpm limiter is nil")
			}

			if limiter.rpd == nil {
				t.Error("rpd limiter is nil")
			}

			if limiter.tpm == nil {
				t.Error("tpm limiter is nil")
			}
		})
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name      string
		options   RateLimiterOptions
		tokens    []int
		wantError bool
		setupWait time.Duration
	}{
		{
			name: "normal operation without tokens",
			options: RateLimiterOptions{
				RPM: 1000,
				RPD: 100000,
				TPM: 10000,
			},
			tokens:    nil,
			wantError: false,
		},
		{
			name: "normal operation with tokens",
			options: RateLimiterOptions{
				RPM: 1000,
				RPD: 100000,
				TPM: 10000,
			},
			tokens:    []int{100},
			wantError: false,
		},
		{
			name: "multiple tokens",
			options: RateLimiterOptions{
				RPM: 1000,
				RPD: 100000,
				TPM: 10000,
			},
			tokens:    []int{50, 25, 25},
			wantError: false,
		},
		{
			name: "context cancellation",
			options: RateLimiterOptions{
				RPM:      1,  // Very slow rate to ensure we can cancel
				RPD:      100,
				TPM:      10,
				BurstRPM: 1, // ensure first call consumes burst only once
			},
			tokens:    nil,
			wantError: true,
			setupWait: 10 * time.Millisecond, // Very short timeout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.options)

			testCtx := ctx
			if tt.setupWait > 0 {
				var cancel context.CancelFunc
				testCtx, cancel = context.WithTimeout(ctx, tt.setupWait)
				defer cancel()
			}

			// For cancellation case, drain initial burst to force blocking
			if tt.wantError && tt.setupWait > 0 && len(tt.tokens) == 0 {
				_ = limiter.Wait(context.Background())
			}

			for _, tokenCount := range tt.tokens {
				err := limiter.Wait(testCtx, tokenCount)
				if tt.wantError {
					if err == nil {
						t.Errorf("Wait() expected error but got none")
					}
					return
				}
				if err != nil {
					t.Errorf("Wait() unexpected error: %v", err)
					return
				}
			}

			// Test without tokens parameter
			err := limiter.Wait(testCtx)
			if tt.wantError {
				if err == nil {
					t.Errorf("Wait() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Wait() unexpected error: %v", err)
			}
		})
	}
}

func TestRateLimiter_RateLimitingBehavior(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a very restrictive rate limiter for testing
	limiter := NewRateLimiter(RateLimiterOptions{
		RPM:      10, // 10 requests per minute = 1 request every 6 seconds
		RPD:      100,
		TPM:      100,
		BurstRPM: 1, // only one immediate request allowed
	})

	start := time.Now()

	// First request should succeed immediately
	err := limiter.Wait(ctx)
	if err != nil {
		t.Fatalf("First Wait() failed: %v", err)
	}

	firstDuration := time.Since(start)
	if firstDuration > 100*time.Millisecond {
		t.Errorf("First request took too long: %v", firstDuration)
	}

	// Second request should be rate limited
	start = time.Now()
	err = limiter.Wait(ctx)
	if err != nil {
		t.Fatalf("Second Wait() failed: %v", err)
	}

	secondDuration := time.Since(start)
	// Should take at least 5.9 seconds (close to 6 seconds for 10 RPM)
	expectedMinDuration := 5900 * time.Millisecond
	if secondDuration < expectedMinDuration {
		t.Errorf("Second request was not properly rate limited. Expected at least %v, got %v", expectedMinDuration, secondDuration)
	}
}

func TestRateLimiter_TokenRateLimiting(t *testing.T) {
ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Create a rate limiter with restrictive token limits
	limiter := NewRateLimiter(RateLimiterOptions{
		RPM:      1000,
		RPD:      100000,
		TPM:      10, // Only 10 tokens per minute = ~0.166 tokens per second
		BurstTPM: 6,  // allow requesting 6 tokens; after first 5, need ~5 more -> ~30s
	})

	start := time.Now()

	// First request with 5 tokens should succeed immediately
	err := limiter.Wait(ctx, 5)
	if err != nil {
		t.Fatalf("First Wait() with tokens failed: %v", err)
	}

	firstDuration := time.Since(start)
	if firstDuration > 100*time.Millisecond {
		t.Errorf("First token request took too long: %v", firstDuration)
	}

	// Second request with 6 tokens should be rate limited (exceeds remaining capacity)
	start = time.Now()
	err = limiter.Wait(ctx, 6)
	if err != nil {
		t.Fatalf("Second Wait() with tokens failed: %v", err)
	}

	secondDuration := time.Since(start)
	// With BurstTPM=5 and TPM=10, requesting 6 tokens requires ~6 tokens to refill (~36s)
	if secondDuration < 30*time.Second {
		t.Errorf("Token rate limiting didn't work as expected. Duration: %v", secondDuration)
	}
}

func TestRateLimiter_InfiniteLimits(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a rate limiter with infinite token limits
	limiter := NewRateLimiter(RateLimiterOptions{
		RPM: 100,
		RPD: 1000,
		TPM: rate.Inf, // Infinite tokens
	})

	start := time.Now()

	// Multiple large token requests should succeed immediately
	for i := 0; i < 10; i++ {
		err := limiter.Wait(ctx, 1000) // Large token count
		if err != nil {
			t.Fatalf("Wait() with infinite tokens failed: %v", err)
		}
	}

	totalDuration := time.Since(start)
	// All requests should complete very quickly due to infinite rate
	if totalDuration > 1*time.Second {
		t.Errorf("Requests with infinite rate took too long: %v", totalDuration)
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	limiter := NewRateLimiter(RateLimiterOptions{
		RPM: 100,
		RPD: 10000,
		TPM: 1000,
	})

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			err := limiter.Wait(ctx, 10)
			if err != nil {
				t.Errorf("Concurrent Wait() %d failed: %v", id, err)
				return
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent test timed out")
		}
	}
}
