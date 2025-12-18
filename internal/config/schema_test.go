package config

import (
	"strings"
	"testing"
)

func TestConfig_Validate_ValidProviders(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{"empty provider (disabled)", "", false},
		{"openrouter", "openrouter", false},
		{"local", "local", false},
		{"OpenRouter uppercase", "OpenRouter", false},
		{"LOCAL uppercase", "LOCAL", false},
		{"invalid provider", "openai", true},
		{"anthropic invalid", "anthropic", true},
		{"unknown", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.LLM.Provider = tt.provider

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "llm.provider") {
					t.Errorf("Error should mention llm.provider, got: %v", err)
				}
			}
		})
	}
}

func TestConfig_Validate_Theme(t *testing.T) {
	tests := []struct {
		name    string
		theme   string
		wantErr bool
	}{
		{"dark", "dark", false},
		{"light", "light", false},
		{"Dark uppercase", "Dark", false},
		{"invalid theme", "blue", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.UI.Theme = tt.theme

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_Temperature(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		wantErr     bool
	}{
		{"zero", 0.0, false},
		{"normal", 0.7, false},
		{"max valid", 2.0, false},
		{"negative", -0.1, true},
		{"too high", 2.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.LLM.Temperature = tt.temperature

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_MaxTokens(t *testing.T) {
	tests := []struct {
		name      string
		maxTokens int
		wantErr   bool
	}{
		{"zero", 0, false},
		{"positive", 1000, false},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.LLM.MaxTokens = tt.maxTokens

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_Timeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
		wantErr bool
	}{
		{"zero", 0, false},
		{"positive", 60, false},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.LLM.Timeout = tt.timeout

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_MaxDiffLines(t *testing.T) {
	tests := []struct {
		name         string
		maxDiffLines int
		wantErr      bool
	}{
		{"zero", 0, false},
		{"positive", 300, false},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.LLM.Privacy.MaxDiffLines = tt.maxDiffLines

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_LocalProviderRequiresBaseURL(t *testing.T) {
	cfg := Default()
	cfg.LLM.Provider = "local"
	cfg.LLM.Local.BaseURL = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for local provider without base_url")
	}
	if !strings.Contains(err.Error(), "base_url") {
		t.Errorf("Error should mention base_url, got: %v", err)
	}
}

func TestConfig_Validate_LocalProviderWithBaseURL(t *testing.T) {
	cfg := Default()
	cfg.LLM.Provider = "local"
	cfg.LLM.Local.BaseURL = "http://localhost:11434"

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDefault_ReturnsValidConfig(t *testing.T) {
	cfg := Default()

	if err := cfg.Validate(); err != nil {
		t.Errorf("Default config should be valid, got error: %v", err)
	}
}

func TestDefault_OpenRouterIsDefault(t *testing.T) {
	cfg := Default()

	if cfg.LLM.Provider != "openrouter" {
		t.Errorf("Default provider = %q, want %q", cfg.LLM.Provider, "openrouter")
	}
}

func TestDefault_ModelFormat(t *testing.T) {
	cfg := Default()

	// OpenRouter models should be in provider/model format
	if !strings.Contains(cfg.LLM.Model, "/") {
		t.Errorf("Default model should be in provider/model format, got: %q", cfg.LLM.Model)
	}
}

func TestOpenRouterConfig_Fields(t *testing.T) {
	cfg := OpenRouterConfig{
		BaseURL:  "https://custom.api.com",
		SiteURL:  "https://myapp.com",
		SiteName: "MyApp",
	}

	if cfg.BaseURL != "https://custom.api.com" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://custom.api.com")
	}
	if cfg.SiteURL != "https://myapp.com" {
		t.Errorf("SiteURL = %q, want %q", cfg.SiteURL, "https://myapp.com")
	}
	if cfg.SiteName != "MyApp" {
		t.Errorf("SiteName = %q, want %q", cfg.SiteName, "MyApp")
	}
}

func TestLLMConfig_HasOpenRouterField(t *testing.T) {
	cfg := Default()

	// Verify OpenRouter config is accessible
	_ = cfg.LLM.OpenRouter.BaseURL
	_ = cfg.LLM.OpenRouter.SiteURL
	_ = cfg.LLM.OpenRouter.SiteName
}
