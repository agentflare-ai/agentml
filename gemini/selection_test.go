package gemini

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComplexityAnalysis(t *testing.T) {
	testCases := []struct {
		prompt   string
		expected string
		reason   string
	}{
		{
			prompt:   "rename variable x to y",
			expected: "simple",
			reason:   "Simple keyword detected",
		},
		{
			prompt:   "format this code properly",
			expected: "simple",
			reason:   "Simple formatting task",
		},
		{
			prompt:   "refactor this function to improve readability",
			expected: "moderate",
			reason:   "Refactoring task",
		},
		{
			prompt:   "implement a new authentication service with JWT tokens",
			expected: "moderate",
			reason:   "Implementation with specific features",
		},
		{
			prompt:   "design microservice architecture for distributed system",
			expected: "complex",
			reason:   "Architecture design task",
		},
		{
			prompt:   "create sophisticated algorithm for performance optimization",
			expected: "complex",
			reason:   "Complex algorithm design",
		},
		{
			prompt:   "",
			expected: "simple",
			reason:   "Empty prompt",
		},
		{
			prompt:   "fix bug",
			expected: "simple",
			reason:   "Simple fix",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.prompt, func(t *testing.T) {
			result := AnalyzeComplexity(tc.prompt)
			assert.Equal(t, tc.expected, result.Level, "Expected complexity level %s but got %s for prompt: %s", tc.expected, result.Level, tc.prompt)
			assert.NotEmpty(t, result.Reason, "Reason should not be empty")
			assert.GreaterOrEqual(t, result.Confidence, 0.0, "Confidence should be non-negative")
			assert.LessOrEqual(t, result.Confidence, 1.0, "Confidence should not exceed 1.0")
		})
	}
}

func TestModelSelectionStrategy(t *testing.T) {
	strategy := NewModelSelectionStrategy()
	ctx := context.Background()

	testCases := []struct {
		prompt        string
		expectedModel ModelName
		description   string
	}{
		{
			prompt:        "rename variable x to y",
			expectedModel: FlashLite,
			description:   "Simple task should select FlashLite",
		},
		{
			prompt:        "refactor this function",
			expectedModel: Flash,
			description:   "Moderate task should select Flash",
		},
		{
			prompt:        "design microservice architecture",
			expectedModel: Pro,
			description:   "Complex task should select Ultra",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := strategy.SelectModel(ctx, tc.prompt)
			assert.Equal(t, tc.expectedModel, result.PrimaryModel, "Expected model %s but got %s", tc.expectedModel, result.PrimaryModel)
			assert.NotEmpty(t, result.FallbackModels, "Should have fallback models")
			assert.NotEmpty(t, result.SelectionReason, "Should have selection reason")
			assert.NotEmpty(t, result.Complexity.Level, "Should have complexity analysis")
		})
	}
}

func TestModelSelectionFallbackChain(t *testing.T) {
	strategy := NewModelSelectionStrategy()
	ctx := context.Background()

	result := strategy.SelectModel(ctx, "design complex architecture")

	// Verify fallback chain doesn't include the primary model
	for _, fallback := range result.FallbackModels {
		assert.NotEqual(t, result.PrimaryModel, fallback, "Fallback chain should not include primary model")
	}

	// Verify fallback chain has reasonable length
	assert.LessOrEqual(t, len(result.FallbackModels), strategy.MaxFallbackAttempts, "Fallback chain should not exceed max attempts")
}

func TestIsRateLimitError(t *testing.T) {
	testCases := []struct {
		err      error
		expected bool
		name     string
	}{
		{
			err:      nil,
			expected: false,
			name:     "nil error",
		},
		{
			err:      &RateLimitError{Model: Flash, Underlying: assert.AnError},
			expected: true,
			name:     "explicit RateLimitError",
		},
		{
			err:      assert.AnError,
			expected: false,
			name:     "generic error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRateLimitError(tc.err)
			assert.Equal(t, tc.expected, result, "Expected %v but got %v for error: %v", tc.expected, result, tc.err)
		})
	}
}

func TestCountKeywordMatches(t *testing.T) {
	keywords := []string{"refactor", "implement", "design"}

	testCases := []struct {
		prompt   string
		expected int
		name     string
	}{
		{
			prompt:   "refactor this code",
			expected: 1,
			name:     "single match",
		},
		{
			prompt:   "refactor and implement new design",
			expected: 3,
			name:     "multiple matches",
		},
		{
			prompt:   "no matching words here",
			expected: 0,
			name:     "no matches",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := countKeywordMatches(tc.prompt, keywords)
			assert.Equal(t, tc.expected, result, "Expected %d matches but got %d", tc.expected, result)
		})
	}
}

func TestContainsMultipleComplexConcepts(t *testing.T) {
	testCases := []struct {
		prompt   string
		expected bool
		name     string
	}{
		{
			prompt:   "simple task",
			expected: false,
			name:     "no complex concepts",
		},
		{
			prompt:   "design architecture",
			expected: false,
			name:     "single complex concept",
		},
		{
			prompt:   "design microservice architecture with performance optimization",
			expected: true,
			name:     "multiple complex concepts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := containsMultipleComplexConcepts(tc.prompt)
			assert.Equal(t, tc.expected, result, "Expected %v but got %v for prompt: %s", tc.expected, result, tc.prompt)
		})
	}
}

func TestNewModelSelectionStrategy(t *testing.T) {
	strategy := NewModelSelectionStrategy()

	// Verify default mappings
	assert.Equal(t, FlashLite, strategy.PreferredModels["simple"])
	assert.Equal(t, Flash, strategy.PreferredModels["moderate"])
	assert.Equal(t, Pro, strategy.PreferredModels["complex"])

	// Verify fallback chain is reasonable
	assert.NotEmpty(t, strategy.FallbackChain)
	assert.Greater(t, strategy.MaxFallbackAttempts, 0)
	assert.LessOrEqual(t, strategy.MaxFallbackAttempts, 5) // Reasonable upper bound
}
