package local

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bui/internal/config"
	"bui/internal/llm/core"
)

// Provider targets an Ollama-compatible streaming endpoint by default.
// Config: [llm.local] base_url = "http://localhost:11434"
type Provider struct {
	baseURL     string
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
		baseURL:     strings.TrimRight(cfg.LLM.Local.BaseURL, "/"),
		model:       cfg.LLM.Model,
		maxTokens:   cfg.LLM.MaxTokens,
		temperature: cfg.LLM.Temperature,
		client:      &http.Client{Timeout: defaultHTTPTimeout},
	}
}

func (p *Provider) Name() string { return "local" }

type reqBody struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options,omitempty"`
}

type respLine struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

func (p *Provider) Stream(ctx context.Context, pr core.Prompt) (<-chan core.Token, <-chan error) {
	out := make(chan core.Token, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if p.baseURL == "" {
			errCh <- fmt.Errorf("llm.local.base_url is empty")
			return
		}

		// Local models often don't have a separate system field, so we prepend it.
		combined := pr.System + "\n\n" + pr.User

		b, err := json.Marshal(reqBody{
			Model:  p.model,
			Prompt: combined,
			Stream: true,
			Options: map[string]any{
				"temperature": firstNonZeroFloat(pr.Temperature, p.temperature),
			},
		})
		if err != nil {
			errCh <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := p.baseURL + "/api/generate"
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
		if err != nil {
			errCh <- err
			return
		}
		req.Header.Set("Content-Type", "application/json")

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
			errCh <- fmt.Errorf("local http %d: %s", resp.StatusCode, buf.String())
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			var r respLine
			if err := json.Unmarshal([]byte(line), &r); err != nil {
				continue
			}
			if r.Error != "" {
				errCh <- fmt.Errorf("local llm error: %s", r.Error)
				return
			}
			if r.Response != "" {
				select {
				case out <- core.Token{Text: r.Response}:
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
			if r.Done {
				return
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

func firstNonZeroFloat(v float64, fallback float64) float64 {
	if v != 0 {
		return v
	}
	return fallback
}
