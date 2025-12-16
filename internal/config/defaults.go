package config

func Default() Config {
	return Config{
		UI: UIConfig{
			Theme:          "dark",
			Accent:         "indigo",
			SyntaxTheme:    "github-dark",
			ShowLineNumbers: true,
		},
		Keys: KeysConfig{
			Quit:    "q",
			Help:    "?",
			Confirm: "enter",
		},
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4.1-mini",
			MaxTokens:   800,
			Temperature: 0.2,
			Stream:      true,
			Privacy: LLMPrivacyConfig{
				SendFullDiff: false,
				MaxDiffLines: 300,
			},
			Local: LocalLLMConfig{
				BaseURL: "http://localhost:11434",
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
