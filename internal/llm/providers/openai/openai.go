package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"bui/internal/config"
	"bui/internal/llm/core"
)

type Provider struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	client      *http.Client
}

// defaultHTTPTimeout is the timeout for establishing HTTP connections.
// The streaming response timeout is controlled by context cancellation.
const defaultHTTPTimeout = 30 * time.Second

func New(cfg config.Config) *Provider {
	return &Provider{
		apiKey:      os.Getenv("OPENAI_API_KEY"),
		model:       cfg.LLM.Model,
		maxTokens:   cfg.LLM.MaxTokens,
		temperature: cfg.LLM.Temperature,
		client:      &http.Client{Timeout: defaultHTTPTimeout},
	}
}

func (p *Provider) Name() string { return "openai" }

type chatReq struct {
	Model       string  `json:"model"`
	Stream      bool    `json:"stream"`
	Messages    []msg   `json:"messages"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (p *Provider) Stream(ctx context.Context, pr core.Prompt) (<-chan core.Token, <-chan error) {
	out := make(chan core.Token, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if p.apiKey == "" {
			errCh <- fmt.Errorf("OPENAI_API_KEY is not set")
			return
		}

		body, err := json.Marshal(chatReq{
			Model:  p.model,
			Stream: true,
			Messages: []msg{
				{Role: "system", Content: pr.System},
				{Role: "user", Content: pr.User},
			},
			MaxTokens:   firstNonZero(pr.MaxTokens, p.maxTokens),
			Temperature: firstNonZeroFloat(pr.Temperature, p.temperature),
		})
		if err != nil {
			errCh <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
		if err != nil {
			errCh <- err
			return
		}
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("Content-Type", "application/json")

		// Streaming can take time; avoid client timeout by using a transport-level approach.
		client := *p.client
		client.Timeout = 0

		resp, err := client.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode >= 400 {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(resp.Body)
			errCh <- fmt.Errorf("openai http %d: %s", resp.StatusCode, buf.String())
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if line == "data: [DONE]" {
				return
			}
			if !bytes.HasPrefix([]byte(line), []byte("data: ")) {
				continue
			}
			payload := line[len("data: "):]
			var c streamChunk
			if err := json.Unmarshal([]byte(payload), &c); err != nil {
				// tolerate non-json noise
				continue
			}
			if len(c.Choices) == 0 {
				continue
			}
			tok := c.Choices[0].Delta.Content
			if tok != "" {
				select {
				case out <- core.Token{Text: tok}:
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
			return
		}

		// Give downstream a moment to render final tokens in some terminals.
		time.Sleep(10 * time.Millisecond)
	}()

	return out, errCh
}

func firstNonZero(v int, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

func firstNonZeroFloat(v float64, fallback float64) float64 {
	if v != 0 {
		return v
	}
	return fallback
}
