package ollama

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/agentml/prompt"
	"github.com/agentflare-ai/go-xmldom"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OllamaNamespaceURI is the XML namespace URI used for Ollama executable elements.
const OllamaNamespaceURI = "github.com/agentflare-ai/agentml/ollama"

// Generate represents an Ollama AI generation executable content element for SCXML.
// It implements the scxml.Executable interface to provide AI generation capabilities
// within SCXML state machines using Ollama models.
//
// The Generate struct maps to XML elements with the following attributes:
//   - model: Specifies the Ollama model to use (e.g., "llama3.2", "codellama")
//   - prompt: The prompt for AI generation
//   - location: Data model location where the generated result should be stored
//   - stream: Whether to use streaming generation (optional, default false)
//   - modelexpr: Dynamic model expression (optional)
//   - promptexpr: Dynamic prompt expression (optional)
type Generate struct {
	xmldom.Element

	// Model specifies the Ollama model to use for generation.
	// Common values include "llama3.2", "llama3.1", "codellama", "mistral", etc.
	Model string `xml:"model,attr"`

	// Prompt contains the prompt for AI generation.
	// This can be a static string or contain data model expressions.
	Prompt string `xml:"prompt,attr"`

	// Location specifies where in the data model to store the generated result.
	// This should be a valid data model location expression.
	Location string `xml:"location,attr"`

	// Stream indicates whether to use streaming generation for real-time responses.
	// When true, responses are delivered progressively as they are generated.
	Stream bool `xml:"stream,attr"`

	// client is the Ollama client for making API calls
	client *Client
}

// Execute implements the scxml.Executable interface for Generate.
// It performs AI generation using the specified Ollama model and prompt,
// then stores the result in the specified data model location.
//
// The execution process:
//  1. Validates that all required attributes are present
//  2. Evaluates the prompt expression using the data model (if needed)
//  3. Calls the Ollama API to generate content
//  4. Stores the generated result in the specified location
//
// Returns an error if generation fails or if required attributes are missing.
func (g *Generate) Execute(ctx context.Context, interpreter agentml.Interpreter) error {
	// Validate required attributes
	modelExpr := string(g.Element.GetAttribute("modelexpr"))
	if g.Model == "" && modelExpr == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Generate element missing required 'model' or 'modelexpr' attribute",
			Data:      map[string]any{"element": "ollama:generate", "line": 0},
			Cause:     fmt.Errorf("generate element missing required 'model' or 'modelexpr' attribute"),
		}
	}

	if g.Location == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Generate element missing required 'location' attribute",
			Data:      map[string]any{"element": "ollama:generate", "line": 0},
			Cause:     fmt.Errorf("generate element missing required 'location' attribute"),
		}
	}

	dataModel := interpreter.DataModel()
	if dataModel == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "No data model available for Ollama generation",
			Data:      map[string]any{"element": "ollama:generate", "line": 0},
			Cause:     fmt.Errorf("no data model available for ollama generation"),
		}
	}

	// Support dynamic modelexpr (evaluated via data model)
	modelName := g.Model
	if me := string(g.Element.GetAttribute("modelexpr")); strings.TrimSpace(me) != "" {
		v, err := dataModel.EvaluateValue(ctx, me)
		if err == nil {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				modelName = s
			}
		}
	}

	// Start OpenTelemetry span for execution tracking
	tracer := otel.Tracer("ollama")
	ctx, span := tracer.Start(ctx, "ollama.generate.execute",
		trace.WithAttributes(
			attribute.String("ollama.model", modelName),
			attribute.String("ollama.location", g.Location),
		),
	)
	defer span.End()

	// Evaluate prompt expression if it contains data model expressions
	promptText, err := g.evaluatePrompt(ctx, interpreter)
	if err != nil {
		span.RecordError(err)
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   fmt.Sprintf("Failed to evaluate prompt expression: %v", err),
			Data:      map[string]any{"element": "ollama:generate", "line": 0},
			Cause:     err,
		}
	}

	// Also support dynamic promptexpr (evaluated via data model)
	if pe := string(g.Element.GetAttribute("promptexpr")); strings.TrimSpace(pe) != "" {
		v, err := dataModel.EvaluateValue(ctx, pe)
		if err == nil {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				if promptText != "" {
					promptText += "\n"
				}
				promptText += s
			}
		}
	}

	// Process child <ollama:prompt> elements as templates
	childPrompts, err := g.processChildPrompts(ctx, interpreter)
	if err != nil {
		span.RecordError(err)
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   fmt.Sprintf("Failed to process child prompt elements: %v", err),
			Data:      map[string]any{"element": "ollama:generate", "line": 0},
			Cause:     err,
		}
	}

	// Combine prompts - attribute prompt first, then child prompts
	finalPrompt := promptText
	if len(childPrompts) > 0 {
		if finalPrompt != "" {
			finalPrompt += "\n" + strings.Join(childPrompts, "\n")
		} else {
			finalPrompt = strings.Join(childPrompts, "\n")
		}
	}

	// Build system instruction from SCXML snapshot (like Gemini does)
	var systemPrompt string
	if doc, err := interpreter.Snapshot(ctx, agentml.SnapshotConfig{ExcludeConfiguration: true, ExcludeData: true}); err == nil {
		// Prune redundant information from snapshot
		prompt.PruneSnapshot(doc)

		slog.Debug("ollama.generate.execute: pruned snapshot ready")
		// Marshal and compress for minimal token usage
		if b, err2 := xmldom.Marshal(doc); err2 == nil {
			systemPrompt = prompt.CompressXML(string(b))
		}
	}

	// TODO: Build tools from available actions when Actions() method is implemented
	// For now, tools must be configured separately if needed
	var ollamaTools []api.Tool
	_ = ollamaTools  // Unused for now
	_ = systemPrompt // Unused for now

	span.SetAttributes(attribute.String("ollama.prompt_length", fmt.Sprintf("%d", len(finalPrompt))))

	// Get or initialize Ollama client
	client := g.client
	if client == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Ollama client not configured. Use SetClient() to configure the client.",
			Data:      map[string]any{"element": "ollama:generate", "line": 0},
			Cause:     fmt.Errorf("ollama client not configured"),
		}
	}

	// Use Chat API with tools instead of Generate
	if len(ollamaTools) > 0 {
		// Build messages with system prompt and user prompt
		messages := []api.Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: finalPrompt,
			},
		}

		// Call Chat with tools
		response, err := client.ChatWithTools(ctx, ModelName(modelName), messages, ollamaTools, g.Stream)
		if err != nil {
			span.RecordError(err)
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   fmt.Sprintf("Failed to generate content: %v", err),
				Data:      map[string]any{"element": "ollama:generate", "line": 0},
				Cause:     err,
			}
		}

		// Process tool calls
		if err := processOllamaToolCalls(ctx, interpreter, response); err != nil {
			span.RecordError(err)
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   fmt.Sprintf("Failed to process tool calls: %v", err),
				Data:      map[string]any{"element": "ollama:generate", "line": 0},
				Cause:     err,
			}
		}
	} else {
		// Fall back to regular generation without tools
		response, err := client.Generate(ctx, ModelName(modelName), finalPrompt, g.Stream)
		if err != nil {
			span.RecordError(err)
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   fmt.Sprintf("Failed to generate content: %v", err),
				Data:      map[string]any{"element": "ollama:generate", "line": 0},
				Cause:     err,
			}
		}

		// Store the generated result in the data model
		if err := dataModel.Assign(ctx, g.Location, response); err != nil {
			span.RecordError(err)
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   fmt.Sprintf("Failed to assign result to location '%s': %v", g.Location, err),
				Data:      map[string]any{"element": "ollama:generate", "line": 0},
				Cause:     err,
			}
		}
	}

	return nil
}

// processOllamaToolCalls processes tool calls from Ollama response and sends corresponding events
func processOllamaToolCalls(ctx context.Context, it agentml.Interpreter, resp *api.ChatResponse) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}

	if len(resp.Message.ToolCalls) == 0 {
		return fmt.Errorf("no tool calls in response")
	}

	for _, toolCall := range resp.Message.ToolCalls {
		name := toolCall.Function.Name
		if !strings.HasPrefix(name, "send_") {
			return fmt.Errorf("unsupported function: %s (only send_* allowed)", name)
		}

		// Extract event name
		evName := strings.TrimPrefix(name, "send_")

		// Build event data from arguments
		var data interface{}
		if toolCall.Function.Arguments != nil {
			// Check for nested data or use the whole arguments
			if d, ok := toolCall.Function.Arguments["data"]; ok {
				data = d
			} else if len(toolCall.Function.Arguments) > 0 {
				// Filter out target and delay if present
				filtered := make(map[string]interface{})
				for k, v := range toolCall.Function.Arguments {
					if k != "target" && k != "delay" {
						filtered[k] = v
					}
				}
				if len(filtered) > 0 {
					data = filtered
				}
			}
		}

		// Send SCXML external event
		ev := &agentml.Event{
			Name: evName,
			Type: agentml.EventTypeExternal,
			Data: data,
		}
		if target, ok := toolCall.Function.Arguments["target"].(string); ok && target != "" {
			ev.Origin = target
		}
		if delay, ok := toolCall.Function.Arguments["delay"].(string); ok && delay != "" {
			ev.Delay = delay
		}

		if err := it.Send(ctx, ev); err != nil {
			return err
		}
	}

	return nil
}

// SetClient sets the Ollama client for this Generate instance.
// This enables dependency injection for testing and configuration.
func (g *Generate) SetClient(client *Client) {
	g.client = client
}

// evaluatePrompt evaluates the prompt attribute using the data model if it contains expressions.
// Returns the evaluated prompt text or the original text if no expressions are present.
func (g *Generate) evaluatePrompt(ctx context.Context, interpreter agentml.Interpreter) (string, error) {
	if g.Prompt == "" {
		return "", nil
	}

	// Check if prompt contains data model expressions (simple heuristic)
	if !strings.Contains(g.Prompt, "${") && !strings.Contains(g.Prompt, "{{") {
		// No expressions, return as-is
		return g.Prompt, nil
	}

	// Use data model to evaluate expressions
	dataModel := interpreter.DataModel()
	if dataModel == nil {
		return g.Prompt, nil // Fallback to literal prompt
	}

	// Try to evaluate as an expression
	result, err := dataModel.EvaluateValue(ctx, g.Prompt)
	if err != nil {
		// If evaluation fails, return the original prompt as fallback
		return g.Prompt, nil
	}

	// Convert result to string
	if str, ok := result.(string); ok {
		return str, nil
	}

	// Convert other types to string representation
	return fmt.Sprintf("%v", result), nil
}

// processChildPrompts processes child <ollama:prompt> elements as Go templates.
// Returns a slice of processed prompt strings.
func (g *Generate) processChildPrompts(ctx context.Context, interpreter agentml.Interpreter) ([]string, error) {
	var prompts []string

	// Get child elements
	children := g.ChildNodes()
	if children.Length() == 0 {
		return prompts, nil
	}

	// Get data model context for template processing
	dataModel := interpreter.DataModel()
	var templateData map[string]any
	if dataModel != nil {
		// Try to get all data model variables as template context
		// This is a simplified approach - in practice you might want to expose
		// specific methods to get the current data model state
		templateData = make(map[string]any)
	}

	for i := uint(0); i < children.Length(); i++ {
		child := children.Item(i)
		element, ok := child.(xmldom.Element)
		if !ok {
			continue
		}

		// Check if this is an ollama:prompt element
		localName := element.LocalName()
		namespaceURI := element.NamespaceURI()

		if string(localName) == "prompt" && (string(namespaceURI) == OllamaNamespaceURI || string(namespaceURI) == "") {
			// Get the text content of the prompt element
			promptContent := string(element.TextContent())
			if promptContent == "" {
				continue
			}

			// Process as Go template
			processedPrompt, err := g.processTemplate(promptContent, templateData)
			if err != nil {
				return nil, fmt.Errorf("failed to process template in prompt element: %w", err)
			}

			prompts = append(prompts, processedPrompt)
		}
	}

	return prompts, nil
}

// processTemplate processes a text string as a Go template with the given data.
func (g *Generate) processTemplate(templateText string, data map[string]any) (string, error) {
	// If no template syntax detected, return as-is
	if !strings.Contains(templateText, "{{") {
		return templateText, nil
	}

	// Create and parse template
	tmpl, err := template.New("prompt").Parse(templateText)
	if err != nil {
		// If template parsing fails, return original text
		return templateText, nil
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// If template execution fails, return original text
		return templateText, nil
	}

	return buf.String(), nil
}

// NewGenerate creates a new Generate executable from an XML element.
// It constructs executable content from the provided xmldom.Element by
// extracting attributes needed for generation.
//
// The function validates that required attributes are present and returns
// an error if the element is malformed or missing required data.
//
// Parameters:
//   - ctx: Context for the operation (currently unused but follows interface)
//   - element: The XML element containing the generation configuration
//
// Returns:
//   - agentml.Executor: A new Generate instance implementing the interface
//   - error: An error if the element is invalid or missing required attributes
func NewGenerate(ctx context.Context, element xmldom.Element) (agentml.Executor, error) {
	// Validate element is not nil
	if element == nil {
		return nil, fmt.Errorf("generate element cannot be nil")
	}

	model := string(element.GetAttribute("model"))
	prompt := string(element.GetAttribute("prompt"))
	location := string(element.GetAttribute("location"))
	stream := string(element.GetAttribute("stream")) == "true"

	// Validate required attributes
	if model == "" {
		return nil, fmt.Errorf("generate element missing required 'model' attribute")
	}

	if location == "" {
		return nil, fmt.Errorf("generate element missing required 'location' attribute")
	}

	// Note: prompt can be empty if content will come from child elements

	return &Generate{
		Element:  element,
		Model:    model,
		Prompt:   prompt,
		Location: location,
		Stream:   stream,
	}, nil
}

// Ensure Generate implements the agentml.Executor interface
var _ agentml.Executor = (*Generate)(nil)
