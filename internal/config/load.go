package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	appName  = "bui"
	fileName = "config.toml"

	// configFileMode restricts config file to owner read/write only since
	// the config may contain sensitive settings or API key references.
	configFileMode = 0o600
)

// LoadOrInit loads config from the standard config path.
// If not found, it writes a default config and returns it.
func LoadOrInit() (Config, string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, "", fmt.Errorf("user config dir: %w", err)
	}
	path := filepath.Join(cfgDir, appName, fileName)

	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, "", fmt.Errorf("stat config: %w", err)
		}
		cfg := Default()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Config{}, "", fmt.Errorf("mkdir config dir: %w", err)
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, configFileMode)
		if err != nil {
			return Config{}, "", fmt.Errorf("create config: %w", err)
		}
		defer func() { _ = f.Close() }()

		enc := toml.NewEncoder(f)
		enc.Indent = ""
		if err := enc.Encode(cfg); err != nil {
			return Config{}, "", fmt.Errorf("write config: %w", err)
		}
		return cfg, path, nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, "", fmt.Errorf("decode config: %w", err)
	}

	// Validate config values
	if err := cfg.Validate(); err != nil {
		return Config{}, "", fmt.Errorf("validate config: %w", err)
	}

	return cfg, path, nil
}
