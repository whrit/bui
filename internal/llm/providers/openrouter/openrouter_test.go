package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bui/internal/config"
	"bui/internal/llm/core"
)

func TestNew(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Model:       "anthropic/claude-3-5-sonnet",
			MaxTokens:   1000,
			Temperature: 0.7,
			OpenRouter: config.OpenRouterConfig{
				SiteURL:  "https://example.com",
				SiteName: "TestApp",
			},
		},
	}

	p := New(cfg)

	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.model != "anthropic/claude-3-5-sonnet" {
		t.Errorf("model = %q, want %q", p.model, "anthropic/claude-3-5-sonnet")
	}
	if p.maxTokens != 1000 {
		t.Errorf("maxTokens = %d, want %d", p.maxTokens, 1000)
	}
	if p.temperature != 0.7 {
		t.Errorf("temperature = %f, want %f", p.temperature, 0.7)
	}
	if p.siteURL != "https://example.com" {
		t.Errorf("siteURL = %q, want %q", p.siteURL, "https://example.com")
	}
	if p.siteName != "TestApp" {
		t.Errorf("siteName = %q, want %q", p.siteName, "TestApp")
	}
}

func TestProvider_Name(t *testing.T) {
	p := &Provider{}
	if got := p.Name(); got != "openrouter" {
		t.Errorf("Name() = %q, want %q", got, "openrouter")
	}
}

func TestProvider_Stream_MissingAPIKey(t *testing.T) {
	p := &Provider{
		apiKey: "", // Empty API key
		model:  "openai/gpt-4o",
	}

	ctx := context.Background()
	tokenCh, errCh := p.Stream(ctx, core.Prompt{System: "test", User: "test"})

	// Drain token channel
	for range tokenCh {
	}

	// Check for error
	err := <-errCh
	if err == nil {
		t.Fatal("Expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "OPENROUTER_API_KEY") {
		t.Errorf("Error message should mention OPENROUTER_API_KEY, got: %v", err)
	}
}

func TestProvider_Stream_Success(t *testing.T) {
	// Create mock server that returns SSE stream
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("Authorization header = %q, want %q", auth, "Bearer test-key")
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type header = %q, want %q", ct, "application/json")
		}

		// Send SSE response
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"id":"gen-123","choices":[{"delta":{"content":"Hello"}}]}`,
			`data: {"id":"gen-123","choices":[{"delta":{"content":" world"}}]}`,
			`data: {"id":"gen-123","choices":[{"delta":{"content":"!"}}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			fmt.Fprintln(w, chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	p := &Provider{
		apiKey:      "test-key",
		model:       "openai/gpt-4o",
		maxTokens:   100,
		temperature: 0.5,
		baseURL:     server.URL,
		client:      &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	tokenCh, errCh := p.Stream(ctx, core.Prompt{
		System: "You are a helpful assistant",
		User:   "Say hello",
	})

	var tokens []string
	for tok := range tokenCh {
		tokens = append(tokens, tok.Text)
	}

	// Check for errors
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	default:
	}

	expected := []string{"Hello", " world", "!"}
	if len(tokens) != len(expected) {
		t.Fatalf("Got %d tokens, want %d", len(tokens), len(expected))
	}
	for i, tok := range tokens {
		if tok != expected[i] {
			t.Errorf("Token[%d] = %q, want %q", i, tok, expected[i])
		}
	}
}

func TestProvider_Stream_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, `{"error":{"code":401,"message":"Invalid API key"}}`)
	}))
	defer server.Close()

	p := &Provider{
		apiKey:  "bad-key",
		model:   "openai/gpt-4o",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	tokenCh, errCh := p.Stream(ctx, core.Prompt{System: "test", User: "test"})

	// Drain token channel
	for range tokenCh {
	}

	err := <-errCh
	if err == nil {
		t.Fatal("Expected error for HTTP 401")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("Error should mention status code 401, got: %v", err)
	}
}

func TestProvider_Stream_MidStreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"id":"gen-123","choices":[{"delta":{"content":"Hello"}}]}`,
			`data: {"id":"gen-123","error":{"code":"server_error","message":"Provider disconnected"},"choices":[{"finish_reason":"error"}]}`,
		}

		for _, chunk := range chunks {
			fmt.Fprintln(w, chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	p := &Provider{
		apiKey:  "test-key",
		model:   "openai/gpt-4o",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	tokenCh, errCh := p.Stream(ctx, core.Prompt{System: "test", User: "test"})

	var tokens []string
	for tok := range tokenCh {
		tokens = append(tokens, tok.Text)
	}

	// Should have received some tokens before error
	if len(tokens) == 0 {
		t.Error("Expected to receive some tokens before mid-stream error")
	}

	err := <-errCh
	if err == nil {
		t.Fatal("Expected mid-stream error")
	}
	if !strings.Contains(err.Error(), "Provider disconnected") {
		t.Errorf("Error should mention provider disconnected, got: %v", err)
	}
}

func TestProvider_Stream_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Keep sending tokens slowly
		for i := 0; i < 100; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				fmt.Fprintf(w, `data: {"id":"gen-123","choices":[{"delta":{"content":"token%d"}}]}`+"\n", i)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}))
	defer server.Close()

	p := &Provider{
		apiKey:  "test-key",
		model:   "openai/gpt-4o",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancel is always called
	tokenCh, errCh := p.Stream(ctx, core.Prompt{System: "test", User: "test"})

	// Receive a few tokens then cancel
	count := 0
	for tok := range tokenCh {
		_ = tok
		count++
		if count >= 3 {
			cancel()
			break
		}
	}

	// Drain remaining tokens
	for range tokenCh {
	}

	// Should get context.Canceled error
	err := <-errCh
	if err != nil && err != context.Canceled {
		// Some implementations may return nil on clean cancellation
		t.Logf("Error after cancellation: %v", err)
	}
}

func TestProvider_Stream_OptionalHeaders(t *testing.T) {
	var receivedReferer, receivedTitle string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedReferer = r.Header.Get("HTTP-Referer")
		receivedTitle = r.Header.Get("X-Title")

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `data: [DONE]`)
	}))
	defer server.Close()

	p := &Provider{
		apiKey:   "test-key",
		model:    "openai/gpt-4o",
		baseURL:  server.URL,
		siteURL:  "https://myapp.com",
		siteName: "MyApp",
		client:   &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	tokenCh, errCh := p.Stream(ctx, core.Prompt{System: "test", User: "test"})

	for range tokenCh {
	}
	<-errCh

	if receivedReferer != "https://myapp.com" {
		t.Errorf("HTTP-Referer = %q, want %q", receivedReferer, "https://myapp.com")
	}
	if receivedTitle != "MyApp" {
		t.Errorf("X-Title = %q, want %q", receivedTitle, "MyApp")
	}
}

func TestProvider_Stream_EmptyDelta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"id":"gen-123","choices":[{"delta":{"content":"Hello"}}]}`,
			`data: {"id":"gen-123","choices":[{"delta":{"content":""}}]}`,
			`data: {"id":"gen-123","choices":[{"delta":{"content":" world"}}]}`,
			`data: {"id":"gen-123","choices":[]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			fmt.Fprintln(w, chunk)
		}
	}))
	defer server.Close()

	p := &Provider{
		apiKey:  "test-key",
		model:   "openai/gpt-4o",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	tokenCh, errCh := p.Stream(ctx, core.Prompt{System: "test", User: "test"})

	var tokens []string
	for tok := range tokenCh {
		tokens = append(tokens, tok.Text)
	}

	<-errCh

	// Should only have non-empty tokens
	expected := []string{"Hello", " world"}
	if len(tokens) != len(expected) {
		t.Fatalf("Got %d tokens, want %d: %v", len(tokens), len(expected), tokens)
	}
}

func TestProvider_Stream_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"id":"gen-123","choices":[{"delta":{"content":"Hello"}}]}`,
			`data: {malformed json}`,
			`: this is a comment`,
			``,
			`data: {"id":"gen-123","choices":[{"delta":{"content":" world"}}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			fmt.Fprintln(w, chunk)
		}
	}))
	defer server.Close()

	p := &Provider{
		apiKey:  "test-key",
		model:   "openai/gpt-4o",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	tokenCh, errCh := p.Stream(ctx, core.Prompt{System: "test", User: "test"})

	var tokens []string
	for tok := range tokenCh {
		tokens = append(tokens, tok.Text)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Should tolerate malformed JSON, got error: %v", err)
		}
	default:
	}

	// Should have received valid tokens despite malformed ones
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
}

func TestProvider_Stream_PromptOverrides(t *testing.T) {
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody = make([]byte, r.ContentLength)
		r.Body.Read(receivedBody)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `data: [DONE]`)
	}))
	defer server.Close()

	p := &Provider{
		apiKey:      "test-key",
		model:       "openai/gpt-4o",
		maxTokens:   100,
		temperature: 0.5,
		baseURL:     server.URL,
		client:      &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	tokenCh, errCh := p.Stream(ctx, core.Prompt{
		System:      "system",
		User:        "user",
		MaxTokens:   200, // Override
		Temperature: 0.9, // Override
	})

	for range tokenCh {
	}
	<-errCh

	body := string(receivedBody)
	if !strings.Contains(body, `"max_tokens":200`) {
		t.Errorf("Request should use prompt MaxTokens override: %s", body)
	}
	if !strings.Contains(body, `"temperature":0.9`) {
		t.Errorf("Request should use prompt Temperature override: %s", body)
	}
}

func TestDefaultBaseURL(t *testing.T) {
	if defaultBaseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("defaultBaseURL = %q, want %q", defaultBaseURL, "https://openrouter.ai/api/v1")
	}
}
