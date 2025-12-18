package config

func Default() Config {
	return Config{
		UI: UIConfig{
			Theme:           "dark",
			Accent:          "indigo",
			SyntaxTheme:     "github-dark",
			ShowLineNumbers: true,
		},
		Keys: KeysConfig{
			Quit:    "q",
			Help:    "?",
			Confirm: "enter",
		},
		LLM: LLMConfig{
			Provider:    "openrouter",
			Model:       "anthropic/claude-3-5-sonnet",
			MaxTokens:   800,
			Temperature: 0.2,
			Stream:      true,
			Timeout:     60,
			Privacy: LLMPrivacyConfig{
				SendFullDiff: false,
				MaxDiffLines: 300,
			},
			Local: LocalLLMConfig{
				BaseURL: "http://localhost:11434",
			},
			OpenRouter: OpenRouterConfig{
				// BaseURL defaults to https://openrouter.ai/api/v1
				SiteURL:  "",
				SiteName: "bui",
			},
		},
		Git: GitConfig{
			DefaultBase: "main",
		},
		Behavior: BehaviorConfig{
			ConfirmMerge: true,
			AutoRefresh:  true,
		},
	}
}
