package config

type Config struct {
	UI       UIConfig       `toml:"ui"`
	Keys     KeysConfig     `toml:"keys"`
	LLM      LLMConfig      `toml:"llm"`
	Git      GitConfig      `toml:"git"`
	Behavior BehaviorConfig `toml:"behavior"`
}

type UIConfig struct {
	Theme          string `toml:"theme"`         // dark | light
	Accent         string `toml:"accent"`        // indigo | blue | ...
	SyntaxTheme    string `toml:"syntax_theme"`  // github-dark | github | dracula | ...
	ShowLineNumbers bool  `toml:"show_line_numbers"`
}

type KeysConfig struct {
	Quit    string `toml:"quit"`
	Help    string `toml:"help"`
	Confirm string `toml:"confirm"`
}

type LLMConfig struct {
	Provider    string  `toml:"provider"` // openai | anthropic | local
	Model       string  `toml:"model"`
	MaxTokens   int     `toml:"max_tokens"`
	Temperature float64 `toml:"temperature"`
	Stream      bool    `toml:"stream"`

	Privacy LLMPrivacyConfig `toml:"privacy"`
	Local   LocalLLMConfig   `toml:"local"`
}

type LLMPrivacyConfig struct {
	SendFullDiff bool `toml:"send_full_diff"`
	MaxDiffLines int  `toml:"max_diff_lines"`
}

type LocalLLMConfig struct {
	BaseURL string `toml:"base_url"` // e.g. http://localhost:11434
}

type GitConfig struct {
	DefaultBase string `toml:"default_base"`
}

type BehaviorConfig struct {
	ConfirmMerge bool `toml:"confirm_merge"`
	AutoRefresh  bool `toml:"auto_refresh"`
}
