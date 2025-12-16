package anthropic

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
	"bui/internal/llm"
)

type Provider struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	client      *http.Client
}

func New(cfg config.Config) *Provider {
	return &Provider{
		apiKey:      os.Getenv("ANTHROPIC_API_KEY"),
		model:       cfg.LLM.Model,
		maxTokens:   cfg.LLM.MaxTokens,
		temperature: cfg.LLM.Temperature,
		client:      &http.Client{Timeout: 0},
	}
}

func (p *Provider) Name() string { return "anthropic" }

type msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type reqBody struct {
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature,omitempty"`
	System      string  `json:"system,omitempty"`
	Messages    []msg   `json:"messages"`
	Stream      bool    `json:"stream"`
}

type event struct {
	Type  string `json:"type"`
	Delta struct {
		Text string `json:"text"`
	} `json:"delta"`
}

func (p *Provider) Stream(ctx context.Context, pr llm.Prompt) (<-chan llm.Token, <-chan error) {
	out := make(chan llm.Token, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if p.apiKey == "" {
			errCh <- fmt.Errorf("ANTHROPIC_API_KEY is not set")
			return
		}

		b, _ := json.Marshal(reqBody{
			Model:       p.model,
			MaxTokens:   firstNonZero(pr.MaxTokens, p.maxTokens),
			Temperature: firstNonZeroFloat(pr.Temperature, p.temperature),
			System:      pr.System,
			Messages:    []msg{{Role: "user", Content: pr.User}},
			Stream:      true,
		})

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(b))
		if err != nil {
			errCh <- err
			return
		}
		req.Header.Set("x-api-key", p.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("content-type", "application/json")

		client := *p.client
		client.Timeout = 0
		resp, err := client.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(resp.Body)
			errCh <- fmt.Errorf("anthropic http %d: %s", resp.StatusCode, buf.String())
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			// Anthropic streams as SSE-like lines starting with "data:"
			if line == "data: [DONE]" {
				return
			}
			if !bytes.HasPrefix([]byte(line), []byte("data: ")) {
				continue
			}
			payload := line[len("data: "):]
			var ev event
			if err := json.Unmarshal([]byte(payload), &ev); err != nil {
				continue
			}
			if ev.Type == "content_block_delta" && ev.Delta.Text != "" {
				select {
				case out <- llm.Token{Text: ev.Delta.Text}:
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
