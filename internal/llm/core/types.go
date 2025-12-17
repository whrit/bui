// Package core provides shared types for LLM providers.
// This package exists to break the import cycle between internal/llm and its providers.
package core

import "context"

// Prompt represents an LLM prompt with system and user messages.
type Prompt struct {
	System      string
	User        string
	MaxTokens   int
	Temperature float64
}

// Token represents a single streaming token from an LLM response.
type Token struct {
	Text string
}

// Provider defines the interface for LLM providers.
type Provider interface {
	Name() string
	Stream(ctx context.Context, p Prompt) (<-chan Token, <-chan error)
}
