package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file.
// If the file doesn't exist, it returns nil (no error).
// Existing environment variables are NOT overwritten.
func LoadEnv(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File doesn't exist, that's okay
	}
	return godotenv.Load(path)
}

// LoadEnvFiles loads environment variables from multiple .env files.
// Files are loaded in order; earlier files take precedence since
// godotenv.Load does not overwrite existing variables.
// Missing files are silently ignored.
func LoadEnvFiles(paths ...string) error {
	for _, path := range paths {
		if err := LoadEnv(path); err != nil {
			return err
		}
	}
	return nil
}

// FindEnvFile looks for a .env file in the given directory.
// Returns the path if found, empty string otherwise.
func FindEnvFile(dir string) string {
	path := filepath.Join(dir, ".env")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

// LoadEnvFromWorkingDir loads .env files from the current working directory.
// It looks for .env.local (higher priority) and .env (base config).
// Missing files are silently ignored.
// Existing environment variables are NOT overwritten.
func LoadEnvFromWorkingDir() error {
	cwd, err := os.Getwd()
	if err != nil {
		return nil // Can't get cwd, just skip
	}

	// Load in priority order: .env.local first, then .env
	// godotenv.Load does not overwrite, so higher priority files should load first
	return LoadEnvFiles(
		filepath.Join(cwd, ".env.local"),
		filepath.Join(cwd, ".env"),
	)
}
