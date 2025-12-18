package llm

import (
	"bui/internal/config"
	"bui/internal/llm/providers/local"
	"bui/internal/llm/providers/openrouter"
	"fmt"
)

func NewProvider(cfg config.Config) (Provider, error) {
	switch cfg.LLM.Provider {
	case "openrouter":
		return openrouter.New(cfg), nil
	case "local":
		return local.New(cfg), nil
	default:
		return nil, fmt.Errorf("unknown llm provider: %q", cfg.LLM.Provider)
	}
}
