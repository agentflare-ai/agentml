package stdin

import (
	"context"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
)

func Loader() agentml.NamespaceLoader {
	return func(ctx context.Context, itp agentml.Interpreter, doc xmldom.Document) (agentml.Namespace, error) {
		return &Namespace{itp: itp}, nil
	}
}
