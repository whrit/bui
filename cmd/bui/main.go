package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"bui/internal/app/root"
	"bui/internal/config"
)

func main() {
	// Load environment variables from .env files (if present)
	// Priority: .env.local > .env
	// Existing env vars are NOT overwritten
	if err := config.LoadEnvFromWorkingDir(); err != nil {
		fmt.Fprintln(os.Stderr, "warning: failed to load .env:", err)
	}

	cfg, cfgPath, err := config.LoadOrInit()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}
	m := root.New(cfg, cfgPath)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "run error:", err)
		os.Exit(1)
	}
}
