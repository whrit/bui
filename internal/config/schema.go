package config

import (
	"fmt"
	"strings"
)

// Valid theme options
var validThemes = map[string]bool{
	"dark":  true,
	"light": true,
}

// Valid LLM provider options
var validProviders = map[string]bool{
	"":           true, // disabled
	"openrouter": true,
	"local":      true,
}

// Validate checks that config values are within acceptable ranges.
// Returns an error if any value is invalid.
func (c Config) Validate() error {
	// Validate theme
	theme := strings.ToLower(c.UI.Theme)
	if !validThemes[theme] {
		return fmt.Errorf("invalid ui.theme %q: must be 'dark' or 'light'", c.UI.Theme)
	}

	// Validate LLM provider
	provider := strings.ToLower(c.LLM.Provider)
	if !validProviders[provider] {
		return fmt.Errorf("invalid llm.provider %q: must be 'openrouter', 'local', or empty", c.LLM.Provider)
	}

	// Validate MaxTokens
	if c.LLM.MaxTokens < 0 {
		return fmt.Errorf("invalid llm.max_tokens %d: must be non-negative", c.LLM.MaxTokens)
	}

	// Validate Temperature (typical range 0.0 to 2.0)
	if c.LLM.Temperature < 0 || c.LLM.Temperature > 2 {
		return fmt.Errorf("invalid llm.temperature %.2f: must be between 0.0 and 2.0", c.LLM.Temperature)
	}

	// Validate Timeout
	if c.LLM.Timeout < 0 {
		return fmt.Errorf("invalid llm.timeout %d: must be non-negative", c.LLM.Timeout)
	}

	// Validate MaxDiffLines
	if c.LLM.Privacy.MaxDiffLines < 0 {
		return fmt.Errorf("invalid llm.privacy.max_diff_lines %d: must be non-negative", c.LLM.Privacy.MaxDiffLines)
	}

	// Validate local provider has base URL
	if provider == "local" && c.LLM.Local.BaseURL == "" {
		return fmt.Errorf("llm.local.base_url is required when llm.provider is 'local'")
	}

	return nil
}

type Config struct {
	UI       UIConfig       `toml:"ui"`
	Keys     KeysConfig     `toml:"keys"`
	LLM      LLMConfig      `toml:"llm"`
	Git      GitConfig      `toml:"git"`
	Behavior BehaviorConfig `toml:"behavior"`
}

type UIConfig struct {
	Theme           string `toml:"theme"`        // dark | light
	Accent          string `toml:"accent"`       // indigo | blue | ...
	SyntaxTheme     string `toml:"syntax_theme"` // github-dark | github | dracula | ...
	ShowLineNumbers bool   `toml:"show_line_numbers"`
}

type KeysConfig struct {
	Quit    string `toml:"quit"`
	Help    string `toml:"help"`
	Confirm string `toml:"confirm"`
}

type LLMConfig struct {
	Provider    string  `toml:"provider"` // openrouter | local
	Model       string  `toml:"model"`
	MaxTokens   int     `toml:"max_tokens"`
	Temperature float64 `toml:"temperature"`
	Stream      bool    `toml:"stream"`
	Timeout     int     `toml:"timeout"` // Timeout in seconds for LLM generation (default: 60)

	Privacy    LLMPrivacyConfig `toml:"privacy"`
	Local      LocalLLMConfig   `toml:"local"`
	OpenRouter OpenRouterConfig `toml:"openrouter"`
}

type LLMPrivacyConfig struct {
	SendFullDiff bool `toml:"send_full_diff"`
	MaxDiffLines int  `toml:"max_diff_lines"`
}

type LocalLLMConfig struct {
	BaseURL string `toml:"base_url"` // e.g. http://localhost:11434
}

type OpenRouterConfig struct {
	BaseURL  string `toml:"base_url"`  // Optional: override API endpoint (default: https://openrouter.ai/api/v1)
	SiteURL  string `toml:"site_url"`  // Optional: your app URL for OpenRouter rankings
	SiteName string `toml:"site_name"` // Optional: your app name for OpenRouter rankings
}

type GitConfig struct {
	DefaultBase string `toml:"default_base"`
}

type BehaviorConfig struct {
	ConfirmMerge bool `toml:"confirm_merge"`
	AutoRefresh  bool `toml:"auto_refresh"`
}
