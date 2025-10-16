package stdin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
	"go.opentelemetry.io/otel"
)

const NamespaceURI = "github.com/agentflare-ai/agentml/stdin"

type Namespace struct {
	itp agentml.Interpreter
}

func (n *Namespace) URI() string { return NamespaceURI }

func (n *Namespace) Unload(ctx context.Context) error { return nil }

func (n *Namespace) Handle(ctx context.Context, el xmldom.Element) (bool, error) {
	if el == nil {
		return false, fmt.Errorf("stdin: element cannot be nil")
	}
	local := strings.ToLower(string(el.LocalName()))
	switch local {
	case "read":
		exe := &readExec{Element: el}
		return true, exe.Execute(ctx, n.itp)
	default:
		return false, nil
	}
}

var _ agentml.Namespace = (*Namespace)(nil)

// readExec implements the <stdin:read> executable element
type readExec struct {
	xmldom.Element
}

func (e *readExec) Execute(ctx context.Context, interp agentml.Interpreter) error {
	tr := otel.Tracer("stdin")
	ctx, span := tr.Start(ctx, "stdin.read")
	defer span.End()

	dm := interp.DataModel()
	if dm == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "No data model available for stdin",
			Data:      map[string]any{"element": "read"},
			Cause:     fmt.Errorf("no datamodel"),
		}
	}

	// Get the location/dataid attribute where we should store the result
	loc := string(e.Element.GetAttribute("location"))
	if loc == "" {
		loc = string(e.Element.GetAttribute("dataid"))
	}
	if strings.TrimSpace(loc) == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "stdin:read requires location or dataid attribute",
			Data:      map[string]any{"element": "read"},
			Cause:     fmt.Errorf("missing location"),
		}
	}

	// Get optional prompt attribute
	prompt := string(e.Element.GetAttribute("prompt"))
	if prompt != "" {
		// If promptexpr is specified, evaluate it
		if promptExpr := string(e.Element.GetAttribute("promptexpr")); promptExpr != "" {
			val, err := dm.EvaluateValue(ctx, promptExpr)
			if err != nil {
				return &agentml.PlatformError{
					EventName: "error.execution",
					Message:   "Failed to evaluate promptexpr",
					Data:      map[string]any{"element": "read", "promptexpr": promptExpr},
					Cause:     err,
				}
			}
			if s, ok := val.(string); ok {
				prompt = s
			}
		}
		// Print the prompt to stderr so it doesn't interfere with stdin/stdout
		fmt.Fprint(os.Stderr, prompt)
	}

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return &agentml.PlatformError{
				EventName: "error.execution",
				Message:   "Failed to read from stdin",
				Data:      map[string]any{"element": "read"},
				Cause:     err,
			}
		}
		// EOF reached, store null to distinguish from empty input
		return dm.SetVariable(ctx, loc, nil)
	}

	input := scanner.Text()

	// Store the result in the data model
	if err := dm.SetVariable(ctx, loc, input); err != nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Failed to store stdin input",
			Data:      map[string]any{"element": "read", "location": loc},
			Cause:     err,
		}
	}

	return nil
}
