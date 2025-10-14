package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/go-xmldom"
)

// Loader returns a NamespaceLoader for the memory namespace.
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

func (n *ns) URI() string { return MemoryNamespaceURI }

func (n *ns) Unload(ctx context.Context) error { return nil }

func (n *ns) Handle(ctx context.Context, el xmldom.Element) (bool, error) {
	if el == nil {
		return false, fmt.Errorf("memory: element cannot be nil")
	}
	local := strings.ToLower(string(el.LocalName()))
	switch local {
	case "close", "put", "get", "delete", "copy", "move", "query",
		"kvtruncate", "exec", "begin", "commit", "rollback", "savepoint", "release",
		"sql", "embed", "upsertvector", "search", "deletevector", "vectorindex",
		"addnode", "addedge", "getnode", "getedge", "deletenode", "deleteedge",
		"neighbors", "getneighbors", "graphpath", "graphtruncate", "graphquery":
		exe := &memExec{Element: el, deps: n.deps, local: local}
		return true, exe.Execute(ctx, n.itp)
	case "graph":
		exe := &graphExec{Element: el, deps: n.deps}
		return true, exe.Execute(ctx, n.itp)
	default:
		return false, nil
	}
}
