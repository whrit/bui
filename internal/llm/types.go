package llm

import "context"

type Prompt struct {
	System  string
	User    string
	MaxTokens int
	Temperature float64
}

type Token struct {
	Text string
}

type Provider interface {
	Name() string
	Stream(ctx context.Context, p Prompt) (<-chan Token, <-chan error)
}
