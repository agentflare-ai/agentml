package ollama

import (
	"context"
	"fmt"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
)

// Loader returns a NamespaceLoader for the Ollama namespace.
// It closes over DI deps (Ollama client) and the interpreter.
func Loader(deps *Deps) agentml.NamespaceLoader {
	return func(ctx context.Context, itp agentml.Interpreter, doc xmldom.Document) (agentml.Namespace, error) {
		return &ns{itp: itp, deps: deps}, nil
	}
}

type ns struct {
	itp  agentml.Interpreter
	deps *Deps
}

var _ agentml.Namespace = (*ns)(nil)

func (n *ns) URI() string { return OllamaNamespaceURI }

func (n *ns) Unload(ctx context.Context) error { return nil }

func (n *ns) Handle(ctx context.Context, el xmldom.Element) (bool, error) {
	if el == nil {
		return false, fmt.Errorf("ollama: element cannot be nil")
	}
	switch string(el.LocalName()) {
	case "generate":
		exec, err := NewGenerate(ctx, el)
		if err != nil {
			return true, err
		}
		g := exec.(*Generate)
		if n.deps != nil {
			g.SetClient(n.deps.Client)
		}
		return true, g.Execute(ctx, n.itp)
	default:
		return false, nil
	}
}
