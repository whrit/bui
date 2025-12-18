package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv_FileExists(t *testing.T) {
	// Create a temp .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	content := "TEST_BUI_VAR=hello_world\nTEST_BUI_VAR2=value2"
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Clear any existing value
	os.Unsetenv("TEST_BUI_VAR")
	os.Unsetenv("TEST_BUI_VAR2")

	// Load the .env file
	err := LoadEnv(envFile)
	if err != nil {
		t.Fatalf("LoadEnv failed: %v", err)
	}

	// Verify variables are set
	if got := os.Getenv("TEST_BUI_VAR"); got != "hello_world" {
		t.Errorf("TEST_BUI_VAR = %q, want %q", got, "hello_world")
	}
	if got := os.Getenv("TEST_BUI_VAR2"); got != "value2" {
		t.Errorf("TEST_BUI_VAR2 = %q, want %q", got, "value2")
	}

	// Cleanup
	os.Unsetenv("TEST_BUI_VAR")
	os.Unsetenv("TEST_BUI_VAR2")
}

func TestLoadEnv_FileNotExists(t *testing.T) {
	// Try to load a non-existent file
	err := LoadEnv("/nonexistent/path/.env")

	// Should NOT return an error - missing .env is okay
	if err != nil {
		t.Errorf("LoadEnv should not error on missing file, got: %v", err)
	}
}

func TestLoadEnv_DoesNotOverwriteExisting(t *testing.T) {
	// Create a temp .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	content := "TEST_BUI_EXISTING=from_file"
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Set existing value
	os.Setenv("TEST_BUI_EXISTING", "from_env")

	// Load the .env file
	err := LoadEnv(envFile)
	if err != nil {
		t.Fatalf("LoadEnv failed: %v", err)
	}

	// Verify existing value is NOT overwritten
	if got := os.Getenv("TEST_BUI_EXISTING"); got != "from_env" {
		t.Errorf("TEST_BUI_EXISTING = %q, want %q (should not overwrite)", got, "from_env")
	}

	// Cleanup
	os.Unsetenv("TEST_BUI_EXISTING")
}

func TestLoadEnvFiles_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .env file
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("VAR1=base\nVAR2=base"), 0600); err != nil {
		t.Fatalf("Failed to create .env: %v", err)
	}

	// Create .env.local file (higher priority)
	envLocalFile := filepath.Join(tmpDir, ".env.local")
	if err := os.WriteFile(envLocalFile, []byte("VAR1=local"), 0600); err != nil {
		t.Fatalf("Failed to create .env.local: %v", err)
	}

	// Clear any existing values
	os.Unsetenv("VAR1")
	os.Unsetenv("VAR2")

	// Load files in order: .env.local first, then .env
	// godotenv.Load does NOT overwrite, so load higher priority first
	err := LoadEnvFiles(envLocalFile, envFile)
	if err != nil {
		t.Fatalf("LoadEnvFiles failed: %v", err)
	}

	// VAR1 should be from .env.local (loaded first)
	if got := os.Getenv("VAR1"); got != "local" {
		t.Errorf("VAR1 = %q, want %q", got, "local")
	}

	// VAR2 should be from .env (only defined there)
	if got := os.Getenv("VAR2"); got != "base" {
		t.Errorf("VAR2 = %q, want %q", got, "base")
	}

	// Cleanup
	os.Unsetenv("VAR1")
	os.Unsetenv("VAR2")
}

func TestFindEnvFile_InCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("TEST=1"), 0600); err != nil {
		t.Fatalf("Failed to create .env: %v", err)
	}

	found := FindEnvFile(tmpDir)
	if found != envFile {
		t.Errorf("FindEnvFile() = %q, want %q", found, envFile)
	}
}

func TestFindEnvFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	found := FindEnvFile(tmpDir)
	if found != "" {
		t.Errorf("FindEnvFile() = %q, want empty string", found)
	}
}
