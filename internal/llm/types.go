// Package llm provides LLM integration for bui.
package llm

import (
	"bui/internal/llm/core"
)

// Re-export types from core package for convenience.
// This allows consumers to import just "bui/internal/llm" instead of "bui/internal/llm/core".
type (
	Prompt   = core.Prompt
	Token    = core.Token
	Provider = core.Provider
)
