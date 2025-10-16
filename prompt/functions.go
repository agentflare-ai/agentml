package prompt

import (
	"github.com/agentflare-ai/go-jsonschema"
	"github.com/agentflare-ai/go-xmldom"
)

// SendFunction represents a send_* function declaration for LLMs
type SendFunction struct {
	Name        string
	Description string
	Schema      *jsonschema.Schema
}

// BuildSendFunctions builds function declarations for send events from available transitions.
// This is used by multiple LLM integrations (Gemini, Ollama, etc.) to create
// consistent function declarations for state machine event sending.
func BuildSendFunctions(transitions []xmldom.Element) []SendFunction {
	var functions []SendFunction
	seen := map[string]struct{}{}

	for _, t := range transitions {
		for _, ev := range t.GetAttribute("event") {
			if string(ev) == "" {
				continue
			}
			if _, ok := seen[string(ev)]; ok {
				continue
			}

			// Create schema for this event
			ps := &jsonschema.Schema{
				Type:       jsonschema.TypeObject,
				Properties: map[string]*jsonschema.Schema{},
			}

			// Attach event-specific data schema if present
			if t.GetAttribute("schema") != "" {
				ps.Properties["data"] = &jsonschema.Schema{Type: jsonschema.TypeObject}
			}

			functions = append(functions, SendFunction{
				Name:        "send_" + string(ev),
				Description: "Send event '" + string(ev) + "' through the SCXML interpreter",
				Schema:      ps,
			})
			seen[string(ev)] = struct{}{}
		}
	}

	return functions
}
