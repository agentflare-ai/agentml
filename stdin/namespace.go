package stdin

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
	"go.opentelemetry.io/otel"
)

const NamespaceURI = "github.com/agentflare-ai/agentml/stdin"

type Namespace struct {
	itp    agentml.Interpreter
	reader *bufio.Reader
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
		return true, n.execRead(ctx, el)
	default:
		return false, nil
	}
}

func (n *Namespace) execRead(ctx context.Context, el xmldom.Element) error {
	tr := otel.Tracer("stdin")
	ctx, span := tr.Start(ctx, "stdin.read")
	defer span.End()

	dm := n.itp.DataModel()
	if dm == nil {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "No data model available for stdin",
			Data:      map[string]any{"element": "read"},
			Cause:     fmt.Errorf("no datamodel"),
		}
	}

	loc := string(el.GetAttribute("location"))
	if loc == "" {
		loc = string(el.GetAttribute("dataid"))
	}
	if strings.TrimSpace(loc) == "" {
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "stdin:read requires location or dataid attribute",
			Data:      map[string]any{"element": "read"},
			Cause:     fmt.Errorf("missing location"),
		}
	}

	prompt := string(el.GetAttribute("prompt"))
	if prompt != "" {
		if promptExpr := string(el.GetAttribute("promptexpr")); promptExpr != "" {
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
		fmt.Fprint(os.Stderr, prompt)
	}

	// Create reader on first use and reuse it to avoid buffering issues
	if n.reader == nil {
		n.reader = bufio.NewReader(os.Stdin)
	}

	input, err := n.reader.ReadString('\n')

	if err != nil {
		if err == io.EOF {
			// EOF - return nil
			return dm.SetVariable(ctx, loc, nil)
		}
		return &agentml.PlatformError{
			EventName: "error.execution",
			Message:   "Failed to read from stdin",
			Data:      map[string]any{"element": "read"},
			Cause:     err,
		}
	}

	// Remove trailing newline
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r") // Handle Windows line endings

	return dm.SetVariable(ctx, loc, input)
}

var _ agentml.Namespace = (*Namespace)(nil)
