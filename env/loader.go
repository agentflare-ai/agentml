package env

import (
	"context"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
)

// Loader returns a NamespaceLoader for the env namespace.
// This allows SCXML documents to read and write environment variables.
//
// Usage in SCXML:
//   <scxml xmlns:env="github.com/agentflare-ai/agentml/env">
//     <env:get name="HOME" location="home_dir" />
//     <env:set name="MY_VAR" value="hello" />
//   </scxml>
func Loader() agentml.NamespaceLoader {
	return func(ctx context.Context, itp agentml.Interpreter, doc xmldom.Document) (agentml.Namespace, error) {
		return &Namespace{itp: itp}, nil
	}
}
