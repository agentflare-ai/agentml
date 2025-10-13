package gemini

import (
	"strings"
)

// TaskComplexity represents the complexity analysis of a task prompt.
type TaskComplexity struct {
	// Level indicates the complexity level: "simple", "moderate", or "complex"
	Level string
	// Reason explains why this complexity level was chosen
	Reason string
	// Confidence indicates the confidence score (0.0 to 1.0)
	Confidence float64
}

// AnalyzeComplexity analyzes a prompt to determine task complexity using simple heuristics.
// It categorizes tasks as simple, moderate, or complex based on keyword matching and prompt length.
func AnalyzeComplexity(prompt string) TaskComplexity {
	if prompt == "" {
		return TaskComplexity{
			Level:      "simple",
			Reason:     "Empty prompt",
			Confidence: 1.0,
		}
	}

	// Convert to lowercase for case-insensitive matching
	lowerPrompt := strings.ToLower(prompt)

	// Define keyword sets for different complexity levels
	simpleKeywords := []string{
		"rename", "format", "fix", "change", "update", "replace", "remove", "add", "get", "set",
		"simple", "basic", "quick", "small", "minor", "trivial", "easy",
	}

	complexKeywords := []string{
		"design", "architecture", "algorithm", "optimize", "performance", "scalability",
		"microservice", "distributed", "system", "framework", "pattern", "infrastructure",
		"security", "encryption", "authentication", "authorization", "complex", "advanced",
		"sophisticated", "comprehensive", "enterprise", "protocol", "api design", "database design",
	}

	moderateKeywords := []string{
		"refactor", "restructure", "implement", "integrate", "extend", "enhance", "improve",
		"feature", "functionality", "module", "component", "service", "middleware", "handler",
		"validation", "testing", "configuration", "deployment", "migration",
	}

	// Count keyword matches
	simpleMatches := countKeywordMatches(lowerPrompt, simpleKeywords)
	moderateMatches := countKeywordMatches(lowerPrompt, moderateKeywords)
	complexMatches := countKeywordMatches(lowerPrompt, complexKeywords)

	// Analyze prompt length
	promptLength := len(prompt)

	// Calculate base scores
	var simpleScore, moderateScore, complexScore float64

	// Length-based scoring
	if promptLength < 50 {
		simpleScore += 0.3
	} else if promptLength < 200 {
		moderateScore += 0.2
	} else {
		complexScore += 0.3
	}

	// Keyword-based scoring
	simpleScore += float64(simpleMatches) * 0.4
	moderateScore += float64(moderateMatches) * 0.4
	complexScore += float64(complexMatches) * 0.5

	// Determine complexity level based on highest score
	maxScore := simpleScore
	level := "simple"
	reason := "Simple keywords and/or short prompt length"

	if moderateScore > maxScore {
		maxScore = moderateScore
		level = "moderate"
		reason = "Moderate complexity keywords suggesting refactoring or feature implementation"
	}

	if complexScore > maxScore {
		maxScore = complexScore
		level = "complex"
		reason = "Complex keywords suggesting architecture, design patterns, or sophisticated algorithms"
	}

	// Special cases that override keyword matching
	if containsMultipleComplexConcepts(lowerPrompt) {
		level = "complex"
		reason = "Multiple complex concepts detected"
		maxScore = 0.9
	}

	// Normalize confidence score (cap at 1.0)
	confidence := maxScore
	if confidence > 1.0 {
		confidence = 1.0
	}

	// Minimum confidence threshold
	if confidence < 0.3 {
		confidence = 0.3
	}

	return TaskComplexity{
		Level:      level,
		Reason:     reason,
		Confidence: confidence,
	}
}

// countKeywordMatches counts how many keywords from the list are found in the prompt.
func countKeywordMatches(prompt string, keywords []string) int {
	matches := 0
	for _, keyword := range keywords {
		if strings.Contains(prompt, keyword) {
			matches++
		}
	}
	return matches
}

// containsMultipleComplexConcepts checks if the prompt contains multiple complex concepts
// that would indicate a very complex task requiring Ultra model capabilities.
func containsMultipleComplexConcepts(prompt string) bool {
	complexConcepts := []string{
		"architecture",
		"design pattern",
		"microservice",
		"distributed system",
		"performance optimization",
		"security implementation",
		"algorithm design",
		"database schema",
		"api specification",
		"system integration",
	}

	conceptCount := 0
	for _, concept := range complexConcepts {
		if strings.Contains(prompt, concept) {
			conceptCount++
		}
	}

	// If 2 or more complex concepts are present, consider it very complex
	return conceptCount >= 2
}