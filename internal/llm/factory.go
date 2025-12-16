package llm

import (
	"bui/internal/config"
	"bui/internal/llm/providers/anthropic"
	"bui/internal/llm/providers/local"
	"bui/internal/llm/providers/openai"
	"fmt"
)

func NewProvider(cfg config.Config) (Provider, error) {
	switch cfg.LLM.Provider {
	case "openai":
		return openai.New(cfg), nil
	case "anthropic":
		return anthropic.New(cfg), nil
	case "local":
		return local.New(cfg), nil
	default:
		return nil, fmt.Errorf("unknown llm provider: %q", cfg.LLM.Provider)
	}
}
