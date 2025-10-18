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

const NamespaceURI = "https://xsd.agentml.dev/stdin"

type Namespace struct {
	itp    agentml.Interpreter
	reader *bufio.Reader
}

func (n *Namespace) URI() string { return NamespaceURI }

func (n *Namespace) Unload(ctx context.Context) error { return nil }

func (n *Namespace) Handle(ctx context.Context, el xmldom.Element) (bool, error) {
	fmt.Println("DEBUG: Handle")
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
	fmt.Println("DEBUG: execRead")
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

	// Get event name (defaults to "stdin.read")
	eventName := string(el.GetAttribute("event"))
	if eventName == "" {
		eventName = "stdin.read"
	}

	// Evaluate prompt if present
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

	// Run stdin read in goroutine, selecting on context cancellation
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		input, err := n.reader.ReadString('\n')
		if err != nil {
			errCh <- err
			return
		}
		// Remove trailing newline
		input = strings.TrimSuffix(input, "\n")
		input = strings.TrimSuffix(input, "\r") // Handle Windows line endings
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		case <-errCh:
			return
		case <-resultCh:
			return
		default:
			resultCh <- input
		}
	}()

	// Select on context cancellation or result
	select {
	case <-ctx.Done():
		// Context cancelled - send error event
		return n.itp.Send(ctx, &agentml.Event{
			Name: "error.execution",
			Type: agentml.EventTypeExternal,
			Data: map[string]any{
				"message": "stdin read cancelled",
				"cause":   ctx.Err().Error(),
			},
		})
	case err := <-errCh:
		if err == io.EOF {
			// EOF - send event with nil data
			return n.itp.Send(ctx, &agentml.Event{
				Name: eventName,
				Type: agentml.EventTypeExternal,
				Data: nil,
			})
		}
		// Read error - send error event
		return n.itp.Send(ctx, &agentml.Event{
			Name: "error.execution",
			Type: agentml.EventTypeExternal,
			Data: map[string]any{
				"message": "Failed to read from stdin",
				"cause":   err.Error(),
			},
		})
	case input := <-resultCh:
		// Success - send event with input data
		return n.itp.Send(ctx, &agentml.Event{
			Name: eventName,
			Type: agentml.EventTypeExternal,
			Data: input,
		})
	}
}

var _ agentml.Namespace = (*Namespace)(nil)
