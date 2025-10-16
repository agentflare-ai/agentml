package gemini

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/agentml/prompt"
	"github.com/agentflare-ai/go-xmldom"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genai"
)

// GeminiNamespaceURI is the XML namespace URI used for Gemini executable elements.
const GeminiNamespaceURI = "github.com/agentflare-ai/agentml/gemini"

// Deps holds dependencies for Gemini executables.
// Aligns with the DI style used across the project (e.g., memory package).
type Deps struct {
	Client *Client
}

// Note: In the Namespace-based architecture, use gemini.Loader(deps) to register
// the gemini namespace with the interpreter. The old ExecFactory-based
// registration has been removed.

// Generate represents a Gemini AI generation executable content element for SCXML.
// It implements the scxml.Executable interface to provide AI generation capabilities
// within SCXML state machines using Google's Gemini AI models.
//
// The Generate struct maps to XML elements with the following attributes:
//   - model: Specifies the Gemini model to use (e.g., "gemini-1.5-flash", "gemini-1.5-pro")
//   - prompt: The prompt or template for AI generation
//   - location: Data model location where the generated result should be stored
//
// Example XML usage:
//
//	<gemini:generate model="gemini-1.5-flash"
//	                 prompt="Generate a greeting message"
//	                 location="greeting" />
type Generate struct {
	xmldom.Element

	// Model specifies the Gemini AI model to use for generation.
	// Common values include "gemini-1.5-flash", "gemini-1.5-pro", etc.
	Model string `xml:"model,attr"`

	ModelExpr string `xml:"modelexpr,attr"`

	// Prompt contains the prompt or template for AI generation.
	// This can be a static string or contain data model expressions.
	Prompt string `xml:"prompt,attr"`

	// Location specifies where in the data model to store the generated result.
	// This should be a valid data model location expression.
	Location string `xml:"location,attr"`

	// Stream indicates whether to use streaming generation for real-time responses.
	// When true, responses are delivered progressively as they are generated.
	Stream bool `xml:"stream,attr"`

	// OnChunk specifies the data model location for handling streaming chunks.
	// Only used when Stream is true. Each chunk is assigned to this location.
	OnChunk string `xml:"onchunk,attr"`

	// AutoSelect enables automatic model selection based on prompt complexity.
	// When true, the model attribute becomes optional and is selected automatically.
	AutoSelect bool `xml:"autoselect,attr"`

	// ComplexityHint provides a hint about task complexity for model selection.
	// Valid values: "simple", "moderate", "complex". Optional.
	ComplexityHint string `xml:"complexity,attr"`

	// client is the Gemini client for making API calls
	client *Client
}

// Execute implements the scxml.Executable interface for Generate.
// It performs AI generation using the specified Gemini model and prompt,
// then stores the result in the specified data model location.
//
// The execution process:
//  1. Validates that all required attributes are present
//  2. Evaluates the prompt expression using the data model (if needed)
//  3. Calls the Gemini AI service to generate content
//  4. Stores the generated result in the specified location
//
// Returns an error if generation fails or if required attributes are missing.
func (g *Generate) Execute(ctx context.Context, interpreter agentml.Interpreter) error {
	// Validate required attributes

	if g.Location == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Generate element missing required 'location' attribute",
			Data:      map[string]any{"element": "gemini:generate", "line": 0},
			Cause:     fmt.Errorf("generate element missing required 'location' attribute"),
		}
	}
	dataModel := interpreter.DataModel()
	if dataModel == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "No data model available for Gemini generation",
			Data:      map[string]any{"element": "gemini:generate", "line": 0},
			Cause:     fmt.Errorf("no data model available for gemini generation"),
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
	tracer := otel.Tracer("gemini")
	ctx, span := tracer.Start(ctx, "gemini.generate.execute",
		trace.WithAttributes(
			attribute.String("gemini.model", modelName),
			attribute.String("gemini.location", g.Location),
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
			Data:      map[string]any{"element": "gemini:generate", "line": 0},
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

	// Process child <gemini:prompt> elements as templates
	childPrompts, err := g.processChildPrompts(ctx, interpreter)
	if err != nil {
		span.RecordError(err)
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   fmt.Sprintf("Failed to process child prompt elements: %v", err),
			Data:      map[string]any{"element": "gemini:generate", "line": 0},
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

	// Always set the system instruction to the SCXML model + runtime snapshot
	cfg := &genai.GenerateContentConfig{}
	real, _ := interpreter.(agentml.Interpreter)
	if real != nil {
		if doc, err := real.Snapshot(ctx, agentml.SnapshotConfig{ExcludeConfiguration: true, ExcludeData: true}); err == nil {
			// Prune redundant information from snapshot
			prompt.PruneSnapshot(doc)

			slog.Info("gemini.generate.execute: pruned snapshot", "snapshot", doc)
			// Marshal without indentation for compact representation
			if b, err2 := xmldom.Marshal(doc); err2 == nil {
				// Further compress by removing unnecessary spaces
				compressed := prompt.CompressXML(string(b))
				cfg.SystemInstruction = &genai.Content{Role: "system", Parts: []*genai.Part{genai.NewPartFromText(compressed)}}

				// Write both versions for comparison (only once per run)
				if _, err := os.Stat("penny-snapshot-pruned.xml"); os.IsNotExist(err) {
					// Write pretty version for readability
					if pretty, _ := xmldom.MarshalIndentWithOptions(doc, "", "  ", true); pretty != nil {
						os.WriteFile("penny-snapshot-pruned.xml", pretty, 0644)
					}
					// Write compressed version to see what LLM gets
					os.WriteFile("penny-snapshot-compressed.xml", []byte(compressed), 0644)
					fmt.Printf("[DEBUG] Snapshots written - Original: %d bytes, Compressed: %d bytes\n", len(b), len(compressed))
				}
			}
		}
	}

	// TODO: Derive send_* tools from available actions when Actions() method is implemented
	// For now, tools must be configured separately if needed
	var fns []*genai.FunctionDeclaration
	_ = fns // Unused for now

	span.SetAttributes(attribute.String("gemini.prompt_length", fmt.Sprintf("%d", len(finalPrompt))))

	// Get or initialize Gemini client
	client := g.client
	if client == nil {
		// Client must be provided via DI (Generate.SetClient or factory wiring)
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Gemini client not configured. Use SetClient() to configure the client.",
			Data:      map[string]any{"element": "gemini:generate", "line": 0},
			Cause:     fmt.Errorf("gemini client not configured"),
		}
	}

	// Generate content using Gemini API. Only use auto-selection when autoselect="true".
	var response *genai.GenerateContentResponse
	if g.AutoSelect {
		var selectionResult *ModelSelectionResult
		response, selectionResult, err = client.GenerateWithAutoSelection(ctx, finalPrompt, cfg)
		_ = selectionResult
	} else {
		content := &genai.Content{Role: "user", Parts: []*genai.Part{genai.NewPartFromText(finalPrompt)}}
		resp, err2 := client.GenerateContent(ctx, ModelName(modelName), []*genai.Content{content}, cfg)
		if err2 != nil {
			err = err2
		} else {
			response = resp
		}
	}
	if err != nil {
		span.RecordError(err)
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   fmt.Sprintf("Failed to generate content: %v", err),
			Data:      map[string]any{"element": "gemini:generate", "line": 0},
			Cause:     err,
		}
	}

	// Log token usage if available
	if response != nil && response.UsageMetadata != nil {
		usage := response.UsageMetadata
		fmt.Printf("[Token Usage] Prompt: %d, Response: %d, Total: %d\n", usage.PromptTokenCount, usage.CandidatesTokenCount, usage.TotalTokenCount)
	}

	// Process send_* function calls only; reject free text
	if err := processSendFunctionCalls(ctx, interpreter, response); err != nil {
		span.RecordError(err)
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   fmt.Sprintf("Failed to process function calls: %v", err),
			Data:      map[string]any{"element": "gemini:generate", "line": 0},
			Cause:     err,
		}
	}

	return nil
}

// toolConfigFromSend limits allowed function names to provided declarations.
func toolConfigFromSend(fns []*genai.FunctionDeclaration) *genai.ToolConfig {
	names := make([]string, 0, len(fns))
	for _, f := range fns {
		names = append(names, f.Name)
	}
	return &genai.ToolConfig{FunctionCallingConfig: &genai.FunctionCallingConfig{Mode: genai.FunctionCallingConfigModeAny, AllowedFunctionNames: names}}
}

// processSendFunctionCalls executes send_* function calls and rejects free text.
func processSendFunctionCalls(ctx context.Context, it agentml.Interpreter, resp *genai.GenerateContentResponse) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	processed := false
	for _, cand := range resp.Candidates {
		if cand.Content == nil {
			continue
		}
		for _, part := range cand.Content.Parts {
			if part.FunctionCall == nil {
				if part.Text != "" {
					return fmt.Errorf("free-form text is not allowed; model must call send_* tools")
				}
				continue
			}
			name := part.FunctionCall.Name
			if !strings.HasPrefix(name, "send_") {
				return fmt.Errorf("unsupported function: %s (only send_* allowed)", name)
			}
			processed = true
			// Build event payload; accept either nested data or flattened
			args := part.FunctionCall.Args
			data := args["data"]
			if data == nil && args != nil {
				filtered := map[string]any{}
				for k, v := range args {
					if k == "target" || k == "delay" {
						continue
					}
					filtered[k] = v
				}
				if len(filtered) > 0 {
					data = filtered
				}
			}
			// Send SCXML external event
			evName := strings.TrimPrefix(name, "send_")
			ev := &agentml.Event{
				Name: evName,
				Type: agentml.EventTypeExternal,
				Data: data,
			}
			if target, ok := args["target"].(string); ok && target != "" {
				ev.Origin = target
			}
			if delay, ok := args["delay"].(string); ok && delay != "" {
				ev.Delay = delay
			}
			if err := it.Send(ctx, ev); err != nil {
				return err
			}
		}
	}
	if !processed {
		return fmt.Errorf("no function calls in response")
	}
	return nil
}

// SetClient sets the Gemini client for this Generate instance.
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

// processChildPrompts processes child <gemini:prompt> elements as Go templates.
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

		// Check if this is a gemini:prompt element
		localName := element.LocalName()
		namespaceURI := element.NamespaceURI()

		if string(localName) == "prompt" && (string(namespaceURI) == "github.com/agentflare-ai/agentml/gemini" || string(namespaceURI) == GeminiNamespaceURI || string(namespaceURI) == "") {
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

// extractTextFromResponse extracts text content from a Gemini API response.
func (g *Generate) extractTextFromResponse(response *genai.GenerateContentResponse) (string, error) {
	if response == nil {
		return "", fmt.Errorf("response is nil")
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in Gemini response")
	}

	candidate := response.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in Gemini response")
	}

	// Combine all text parts
	var resultParts []string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			resultParts = append(resultParts, part.Text)
		}
	}

	if len(resultParts) == 0 {
		return "", fmt.Errorf("no text content in Gemini response")
	}

	return strings.Join(resultParts, ""), nil
}

// generateContent makes the actual API call to Gemini to generate content.
func (g *Generate) generateContent(ctx context.Context, client *Client, model, prompt string, config *genai.GenerateContentConfig) (string, error) {
	// Convert model string to ModelName
	modelName := ModelName(model)

	// Create content for the API call with explicit user role
	content := &genai.Content{Role: "user", Parts: []*genai.Part{genai.NewPartFromText(prompt)}}

	// Make the API call
	response, err := client.GenerateContent(ctx, modelName, []*genai.Content{content}, config)
	if err != nil {
		return "", fmt.Errorf("gemini API call failed: %w", err)
	}

	// Extract text from response
	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in Gemini response")
	}

	candidate := response.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in Gemini response")
	}

	// Combine all text parts
	var resultParts []string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			resultParts = append(resultParts, part.Text)
		}
	}

	if len(resultParts) == 0 {
		return "", fmt.Errorf("no text content in Gemini response")
	}

	return strings.Join(resultParts, ""), nil
}

// generateContentStream handles streaming generation for real-time AI responses.
// It manages the streaming lifecycle, processes chunks as they arrive, and handles
// proper cleanup of goroutines and channels.
func (g *Generate) generateContentStream(ctx context.Context, client *Client, model, prompt string, dataModel agentml.DataModel, config *genai.GenerateContentConfig) error {
	// Convert model string to ModelName
	modelName := ModelName(model)

	// Create content for the API call with explicit user role
	content := &genai.Content{Role: "user", Parts: []*genai.Part{genai.NewPartFromText(prompt)}}

	// Create a channel for streaming responses
	respChan := make(chan *genai.GenerateContentResponse, 1)
	defer close(respChan)

	// Start streaming in a goroutine
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		err := client.StreamGenerate(ctx, modelName, []*genai.Content{content}, config, respChan)
		if err != nil {
			errChan <- err
		}
	}()

	// Accumulate the full response for final storage
	var fullResponseParts []string

	// Process streaming responses
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			if err != nil {
				return fmt.Errorf("streaming generation failed: %w", err)
			}
			// Streaming completed successfully, store final result
			if len(fullResponseParts) > 0 {
				finalResult := strings.Join(fullResponseParts, "")
				if err := dataModel.Assign(ctx, g.Location, finalResult); err != nil {
					return fmt.Errorf("failed to assign final result to location '%s': %w", g.Location, err)
				}
			}
			return nil
		case response, ok := <-respChan:
			if !ok {
				// Channel closed, streaming completed
				continue
			}
			if response == nil {
				continue
			}

			// Process the streaming chunk
			if len(response.Candidates) > 0 {
				candidate := response.Candidates[0]
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						fullResponseParts = append(fullResponseParts, part.Text)

						// If OnChunk is specified, store the chunk
						if g.OnChunk != "" {
							if err := dataModel.Assign(ctx, g.OnChunk, part.Text); err != nil {
								return fmt.Errorf("failed to assign chunk to location '%s': %w", g.OnChunk, err)
							}
						}
					}
				}
			}
		}
	}
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
	onChunk := string(element.GetAttribute("onchunk"))
	autoSelect := string(element.GetAttribute("autoselect")) == "true"
	complexityHint := string(element.GetAttribute("complexity"))
	modelExpr := string(element.GetAttribute("modelexpr"))

	// Validate required attributes - model is optional if auto-selection is enabled
	if modelExpr == "" && model == "" && !autoSelect {
		return nil, fmt.Errorf("generate element missing required 'model' attribute (or set autoselect='true')")
	}

	if location == "" {
		return nil, fmt.Errorf("generate element missing required 'location' attribute")
	}

	// Validate complexity hint if provided
	if complexityHint != "" {
		validHints := map[string]bool{"simple": true, "moderate": true, "complex": true}
		if !validHints[complexityHint] {
			return nil, fmt.Errorf("invalid complexity hint '%s', must be 'simple', 'moderate', or 'complex'", complexityHint)
		}
	}

	// Note: prompt can be empty if content will come from child elements

	return &Generate{
		Element:        element,
		Model:          model,
		ModelExpr:      modelExpr,
		Prompt:         prompt,
		Location:       location,
		Stream:         stream,
		OnChunk:        onChunk,
		AutoSelect:     autoSelect,
		ComplexityHint: complexityHint,
	}, nil
}

// Ensure Generate implements the agentml.Executor interface
var _ agentml.Executor = (*Generate)(nil)

func toolConfigFrom(fns []*genai.FunctionDeclaration) *genai.ToolConfig {
	names := make([]string, 0, len(fns))
	for _, f := range fns {
		names = append(names, f.Name)
	}
	return &genai.ToolConfig{FunctionCallingConfig: &genai.FunctionCallingConfig{Mode: genai.FunctionCallingConfigModeAny, AllowedFunctionNames: names}}
}

// TODO: These functions reference types that don't exist yet in agentml package
// They should be re-enabled when Actions, AvailableTransition, and CancellabelEvent types are added

// func buildFunctionsFromActions(a agentml.Actions) []*genai.FunctionDeclaration {
// 	out := []*genai.FunctionDeclaration{}
// 	if len(a.Raise) > 0 {
// 		out = append(out, makeRaiseFn(a.Raise))
// 	}
// 	if len(a.Send) > 0 {
// 		out = append(out, makeSendFns(a.Send)...)
// 	}
// 	if len(a.Cancel) > 0 {
// 		out = append(out, makeCancelFn(a.Cancel))
// 	}
// 	return out
// }
//
// func makeRaiseFn(actions []agentml.AvailableTransition) *genai.FunctionDeclaration {
// 	return &genai.FunctionDeclaration{
// 		Name:        "Raise",
// 		Description: "Raise an internal event in the SCXML interpreter.",
// 		ParametersJsonSchema: &jsonschema.Schema{
// 			Type: jsonschema.TypeObject,
// 			Properties: map[string]*jsonschema.Schema{
// 				"name": {Type: jsonschema.TypeString},
// 				"data": {Type: jsonschema.TypeObject},
// 			},
// 			Required: []string{"name"},
// 		},
// 	}
// }
//
// func makeSendFns(ts []agentml.AvailableTransition) []*genai.FunctionDeclaration {
// 	fns := []*genai.FunctionDeclaration{}
// 	seen := map[string]struct{}{}
// 	for _, t := range ts {
// 		for _, ev := range t.EventAttrs {
// 			if ev == "" {
// 				continue
// 			}
// 			if _, ok := seen[ev]; ok {
// 				continue
// 			}
// 			ps := &jsonschema.Schema{Type: jsonschema.TypeObject, Properties: map[string]*jsonschema.Schema{"target": {Type: jsonschema.TypeString}, "delay": {Type: jsonschema.TypeString}}}
// 			// Allow either nested data or flattened top-level fields
// 			if t.Schema != nil {
// 				ps.Properties["data"] = t.Schema
// 				// Merge top-level properties as optional aliases
// 				for k, v := range t.Schema.Properties {
// 					ps.Properties[k] = v
// 				}
// 			}
// 			fns = append(fns, &genai.FunctionDeclaration{Name: "send_" + ev, Description: "Send event " + ev + " (args may be under 'data' or flattened at top level)", ParametersJsonSchema: ps})
// 			seen[ev] = struct{}{}
// 		}
// 	}
// 	return fns
// }
//
// func makeCancelFn(cs []agentml.CancellabelEvent) *genai.FunctionDeclaration {
// 	enums := make([]any, 0, len(cs))
// 	for _, c := range cs {
// 		enums = append(enums, c.SendID)
// 	}
// 	return &genai.FunctionDeclaration{
// 		Name:                 "Cancel",
// 		Description:          "Cancel a delayed send operation",
// 		ParametersJsonSchema: &jsonschema.Schema{Type: jsonschema.TypeObject, Properties: map[string]*jsonschema.Schema{"sendId": {Type: jsonschema.TypeString, Enum: enums}}, Required: []string{"sendId"}},
// 	}
// }

func processFunctionCalls(ctx context.Context, it agentml.Interpreter, resp *genai.GenerateContentResponse, fns []*genai.FunctionDeclaration, allowText bool, location string) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	dm := it.DataModel()
	processed := false
	for _, cand := range resp.Candidates {
		if cand.Content == nil {
			continue
		}
		for _, part := range cand.Content.Parts {
			if part.FunctionCall == nil {
				if part.Text != "" && allowText {
					if location != "" && dm != nil {
						_ = dm.Assign(ctx, location, part.Text)
					}
					continue
				}
				return fmt.Errorf("non-function content is not allowed; model must return function calls only")
			}
			processed = true
			name := part.FunctionCall.Name
			if strings.HasPrefix(name, "send_") {
				evName := strings.TrimPrefix(name, "send_")
				return handleSendCall(ctx, it, evName, part.FunctionCall.Args)
			}
			switch name {
			case "Raise":
				return handleRaiseCall(ctx, it, part.FunctionCall.Args)
			case "Cancel":
				return handleCancelCall(ctx, it, part.FunctionCall.Args)
			default:
				return fmt.Errorf("unsupported function: %s", name)
			}
		}
	}
	if !processed {
		return fmt.Errorf("no function calls in response")
	}
	return nil
}

func handleRaiseCall(ctx context.Context, it agentml.Interpreter, args map[string]any) error {
	name, _ := args["name"].(string)
	data := args["data"]
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("Raise.name required")
	}
	ev := &agentml.Event{
		Name: name,
		Type: agentml.EventTypeInternal,
		Data: data,
	}
	it.Raise(ctx, ev)
	return nil
}

func handleSendCall(ctx context.Context, it agentml.Interpreter, eventName string, args map[string]any) error {
	data := args["data"]
	target, _ := args["target"].(string)
	delay, _ := args["delay"].(string)
	ev := &agentml.Event{
		Name: eventName,
		Type: agentml.EventTypeExternal,
		Data: data,
	}
	if target != "" {
		ev.Origin = target
	}
	if delay != "" {
		ev.Delay = delay
	}
	return it.Send(ctx, ev)
}

func handleCancelCall(ctx context.Context, it agentml.Interpreter, args map[string]any) error {
	sendId, _ := args["sendId"].(string)
	if strings.TrimSpace(sendId) == "" {
		return fmt.Errorf("Cancel.sendId required")
	}
	return it.Cancel(ctx, sendId)
}
