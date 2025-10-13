package gemini

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genai"
)

// ModelSelectionStrategy defines the strategy for automatically selecting models based on task complexity.
type ModelSelectionStrategy struct {
	// PreferredModels maps complexity levels to preferred model names
	PreferredModels map[string]ModelName
	// FallbackChain defines the fallback order when preferred models are rate-limited
	FallbackChain []ModelName
	// MaxFallbackAttempts limits the number of fallback attempts
	MaxFallbackAttempts int
}

// NewModelSelectionStrategy creates a new model selection strategy with default mappings.
func NewModelSelectionStrategy() *ModelSelectionStrategy {
	return &ModelSelectionStrategy{
		PreferredModels: map[string]ModelName{
			"simple":   FlashLite,
			"moderate": Flash,
			"complex":  Pro,
		},
		FallbackChain: []ModelName{
			Pro,       // High capability, more available
			Flash,     // Good capability, high availability
			FlashLite, // Fastest, highest availability
		},
		MaxFallbackAttempts: 3,
	}
}

// SelectModel selects the best model for a given task prompt.
// It analyzes complexity and returns the preferred model with fallback options.
func (s *ModelSelectionStrategy) SelectModel(ctx context.Context, prompt string) ModelSelectionResult {
	// Start telemetry span
	tracer := otel.Tracer("gemini")
	ctx, span := tracer.Start(ctx, "gemini.selection.select_model",
		trace.WithAttributes(
			attribute.String("gemini.prompt_length", fmt.Sprintf("%d", len(prompt))),
		),
	)
	defer span.End()

	// Analyze task complexity
	complexity := AnalyzeComplexity(prompt)

	span.SetAttributes(
		attribute.String("gemini.complexity.level", complexity.Level),
		attribute.String("gemini.complexity.reason", complexity.Reason),
		attribute.Float64("gemini.complexity.confidence", complexity.Confidence),
	)

	// Get preferred model for this complexity level
	preferredModel, exists := s.PreferredModels[complexity.Level]
	if !exists {
		// Default to moderate complexity if level not found
		preferredModel = s.PreferredModels["moderate"]
	}

	// Build fallback chain starting with preferred model
	fallbackModels := s.buildFallbackChain(preferredModel)

	span.SetAttributes(
		attribute.String("gemini.selected_model", string(preferredModel)),
		attribute.Int("gemini.fallback_options", len(fallbackModels)-1),
	)

	return ModelSelectionResult{
		PrimaryModel:    preferredModel,
		FallbackModels:  fallbackModels[1:], // Exclude primary model from fallbacks
		Complexity:      complexity,
		SelectionReason: fmt.Sprintf("Selected %s for %s complexity task: %s", preferredModel, complexity.Level, complexity.Reason),
	}
}

// buildFallbackChain builds a fallback chain starting with the preferred model.
func (s *ModelSelectionStrategy) buildFallbackChain(preferredModel ModelName) []ModelName {
	var chain []ModelName

	// Start with preferred model
	chain = append(chain, preferredModel)

	// Add other models from fallback chain, avoiding duplicates
	for _, model := range s.FallbackChain {
		if model != preferredModel && len(chain) < s.MaxFallbackAttempts+1 {
			chain = append(chain, model)
		}
	}

	return chain
}

// ModelSelectionResult contains the result of model selection including fallback options.
type ModelSelectionResult struct {
	// PrimaryModel is the recommended model for the task
	PrimaryModel ModelName
	// FallbackModels provides alternative models if the primary fails
	FallbackModels []ModelName
	// Complexity contains the analyzed task complexity
	Complexity TaskComplexity
	// SelectionReason explains why this model was chosen
	SelectionReason string
}

// RateLimitError represents a rate limiting error that can trigger model fallback.
type RateLimitError struct {
	Model      ModelName
	RetryAfter time.Duration
	Underlying error
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for model %s: %v", e.Model, e.Underlying)
}

// IsRateLimitError checks if an error is a rate limiting error.
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	// Check for explicit RateLimitError type
	if _, ok := err.(*RateLimitError); ok {
		return true
	}

	// Check for common rate limit error messages
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "quota") ||
		strings.Contains(errMsg, "too many requests") ||
		strings.Contains(errMsg, "429")
}

// GenerateWithFallback attempts to generate content with automatic model fallback on rate limits.
func (s *ModelSelectionStrategy) GenerateWithFallback(
	ctx context.Context,
	client *Client,
	prompt string,
	config *genai.GenerateContentConfig,
) (*genai.GenerateContentResponse, *ModelSelectionResult, error) {
	// Start telemetry span
	tracer := otel.Tracer("gemini")
	ctx, span := tracer.Start(ctx, "gemini.selection.generate_with_fallback")
	defer span.End()

	// Select model based on prompt complexity
	selection := s.SelectModel(ctx, prompt)

	span.SetAttributes(
		attribute.String("gemini.primary_model", string(selection.PrimaryModel)),
		attribute.Int("gemini.fallback_count", len(selection.FallbackModels)),
	)

	// Build candidate list from selection, but restrict to client-available models.
	candidates := append([]ModelName{selection.PrimaryModel}, selection.FallbackModels...)
	available := make([]ModelName, 0, len(candidates))
	for _, m := range candidates {
		if _, ok := client.models[m]; ok {
			available = append(available, m)
		}
	}
	// If nothing matched availability, use whatever the client has.
	if len(available) == 0 {
		for m := range client.models {
			available = append(available, m)
		}
	}
	// If client only has a single model, normalize to 3 attempts of the same model (as requested).
	if len(available) == 1 {
		available = []ModelName{available[0], available[0], available[0]}
	}

	var lastErr error
	for i, model := range available {
		span.SetAttributes(attribute.String("gemini.attempt_model", string(model)))

		// Create content for API call
		content := genai.NewContentFromText(prompt, "")

		// Attempt generation with current model
		response, err := client.GenerateContent(ctx, model, []*genai.Content{content}, config)
		if err == nil {
			// Success!
			span.SetAttributes(
				attribute.String("gemini.successful_model", string(model)),
				attribute.Int("gemini.attempts_made", i+1),
			)
			// Align selection.PrimaryModel with the actual model we used
			selection.PrimaryModel = model
			return response, &selection, nil
		}

		lastErr = err

		// Check if this is a rate limit error; otherwise continue to next available.
		if IsRateLimitError(err) {
			span.SetAttributes(
				attribute.String("gemini.rate_limit_model", string(model)),
				attribute.Int("gemini.attempt_number", i+1),
			)
			// keep trying next available
			continue
		}
		// Non-rate-limit error: try next available instead of bailing immediately.
		span.RecordError(err)
	}

	// All attempts failed
	if lastErr == nil {
		lastErr = fmt.Errorf("no available models for auto-selection")
	}
	span.RecordError(lastErr)
	return nil, &selection, lastErr
}
