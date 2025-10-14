package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
)

const NamespaceURI = "github.com/agentflare-ai/agentml/agent"

// Loader returns a NamespaceLoader for the agent namespace.
func Loader() agentml.NamespaceLoader {
	return func(ctx context.Context, itp agentml.Interpreter, doc xmldom.Document) (agentml.Namespace, error) {
		return &ns{itp: itp, doc: doc}, nil
	}
}

type ns struct {
	itp agentml.Interpreter
	doc xmldom.Document
}

var _ agentml.Namespace = (*ns)(nil)

func (n *ns) URI() string { return NamespaceURI }

func (n *ns) Unload(ctx context.Context) error { return nil }

func (n *ns) Handle(ctx context.Context, el xmldom.Element) (bool, error) {
	if el == nil {
		return false, fmt.Errorf("agent: element cannot be nil")
	}

	local := strings.ToLower(string(el.LocalName()))
	switch local {
	case "agent":
		exe := &agentExec{Element: el, doc: n.doc}
		return true, exe.Execute(ctx, n.itp)
	case "transition":
		// Check if this transition has an event:schema attribute
		schema := string(el.GetAttribute("event:schema"))
		if schema != "" {
			// Log the schema for now - validation can be added later
			n.itp.Log(ctx, "agent", fmt.Sprintf("Transition event schema: %s", schema))
		}
		// Return false to let the standard SCXML handler process the transition
		return false, nil
	default:
		return false, nil
	}
}

// agentExec implements the <agent> root element
// This element replaces <scxml> as the document root, adding namespace declarations and datamodel specification
type agentExec struct {
	xmldom.Element
	doc xmldom.Document
}

func (e *agentExec) Execute(ctx context.Context, interp agentml.Interpreter) error {
	// Parse the datamodel attribute
	datamodel := string(e.Element.GetAttribute("datamodel"))

	// Parse use:* attributes for namespace declarations (equivalent to xmlns:*)
	// Format: use:prefix="namespace-uri"
	// Example: use:memory="github.com/agentflare-ai/agentml/memory" is equivalent to xmlns:memory="..."
	namespaces := make(map[string]string)

	attrs := e.Element.Attributes()
	for i := uint(0); i < attrs.Length(); i++ {
		attr := attrs.Item(i)
		name := string(attr.NodeName())
		if prefix, ok := strings.CutPrefix(name, "use:"); ok {
			uri := string(attr.NodeValue())
			namespaces[prefix] = uri
		}
	}

	// Log what we found for debugging
	if datamodel != "" {
		interp.Log(ctx, "agent", fmt.Sprintf("Datamodel: %s", datamodel))
	}

	for prefix, uri := range namespaces {
		interp.Log(ctx, "agent", fmt.Sprintf("Namespace %s: %s", prefix, uri))
	}

	// Process child elements for event:schema attributes on transitions
	e.processEventSchemas(ctx, interp, e.Element)

	// Execute child elements
	// The child elements are standard SCXML content (states, transitions, data, etc.)
	children := e.Element.ChildNodes()
	for i := uint(0); i < children.Length(); i++ {
		child := children.Item(i)
		if childEl, ok := child.(xmldom.Element); ok {
			if err := interp.ExecuteElement(ctx, childEl); err != nil {
				return &agentml.PlatformError{
					EventName: "error.execution",
					Message:   fmt.Sprintf("Failed to execute agent child element: %v", err),
					Data: map[string]any{
						"element":   "agent",
						"child":     childEl.TagName(),
						"datamodel": datamodel,
					},
					Cause: err,
				}
			}
		}
	}

	return nil
}

// processEventSchemas recursively processes all transition elements to collect event:schema attributes
func (e *agentExec) processEventSchemas(ctx context.Context, interp agentml.Interpreter, el xmldom.Element) {
	if el == nil {
		return
	}

	localName := strings.ToLower(string(el.LocalName()))

	// Check if this is a transition element with an event:schema attribute
	if localName == "transition" {
		schema := string(el.GetAttribute("event:schema"))
		eventName := string(el.GetAttribute("event"))

		if schema != "" && eventName != "" {
			interp.Log(ctx, "agent", fmt.Sprintf("Event schema for '%s': %s", eventName, schema))
			// TODO: Store the schema for validation when the event is raised
			// This could be stored in the datamodel or a separate schema registry
		}
	}

	// Recursively process children
	children := el.ChildNodes()
	for i := uint(0); i < children.Length(); i++ {
		child := children.Item(i)
		if childEl, ok := child.(xmldom.Element); ok {
			e.processEventSchemas(ctx, interp, childEl)
		}
	}
}

