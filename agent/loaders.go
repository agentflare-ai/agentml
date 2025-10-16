package agent

import (
	"github.com/agentflare-ai/agentml"
	"github.com/agentflare-ai/agentml/env"
	"github.com/agentflare-ai/agentml/gemini"
	"github.com/agentflare-ai/agentml/memory"
	"github.com/agentflare-ai/agentml/ollama"
	"github.com/agentflare-ai/agentml/stdin"
)

// AllLoaders returns a map of all available AgentML namespace loaders.
// Use this to easily register all standard namespaces with an interpreter.
func AllLoaders() map[string]agentml.NamespaceLoader {
	return map[string]agentml.NamespaceLoader{
		NamespaceURI:              Loader(),
		env.NamespaceURI:          env.Loader(),
		stdin.NamespaceURI:        stdin.Loader(),
		gemini.GeminiNamespaceURI: gemini.Loader(nil),
		memory.MemoryNamespaceURI: memory.Loader(nil),
		ollama.OllamaNamespaceURI: ollama.Loader(nil),
	}
}
