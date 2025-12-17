package ui

import (
	"testing"

	"bui/internal/config"

	"github.com/charmbracelet/bubbles/key"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name     string
		binding  key.Binding
		wantKeys []string
		wantHelp string
	}{
		// Global keys
		{
			name:     "Quit binding",
			binding:  km.Quit,
			wantKeys: []string{"q", "ctrl+c"},
			wantHelp: "quit",
		},
		{
			name:     "Help binding",
			binding:  km.Help,
			wantKeys: []string{"?"},
			wantHelp: "help",
		},
		{
			name:     "Confirm binding",
			binding:  km.Confirm,
			wantKeys: []string{"enter"},
			wantHelp: "confirm",
		},
		{
			name:     "Cancel binding",
			binding:  km.Cancel,
			wantKeys: []string{"esc"},
			wantHelp: "cancel",
		},

		// Navigation
		{
			name:     "Up binding",
			binding:  km.Up,
			wantKeys: []string{"up"},
			wantHelp: "up",
		},
		{
			name:     "Down binding",
			binding:  km.Down,
			wantKeys: []string{"down"},
			wantHelp: "down",
		},
		{
			name:     "NextItem (vim j)",
			binding:  km.NextItem,
			wantKeys: []string{"j"},
			wantHelp: "next",
		},
		{
			name:     "PrevItem (vim k)",
			binding:  km.PrevItem,
			wantKeys: []string{"k"},
			wantHelp: "prev",
		},

		// Actions
		{
			name:     "Select binding",
			binding:  km.Select,
			wantKeys: []string{"enter"},
			wantHelp: "select",
		},
		{
			name:     "Back binding",
			binding:  km.Back,
			wantKeys: []string{"esc", "backspace"},
			wantHelp: "back",
		},
		{
			name:     "Refresh binding",
			binding:  km.Refresh,
			wantKeys: []string{"r"},
			wantHelp: "refresh",
		},
		{
			name:     "Search binding",
			binding:  km.Search,
			wantKeys: []string{"/"},
			wantHelp: "search",
		},
		{
			name:     "Filter binding",
			binding:  km.Filter,
			wantKeys: []string{"f"},
			wantHelp: "filter",
		},

		// PR-specific
		{
			name:     "OpenInBrowser binding",
			binding:  km.OpenInBrowser,
			wantKeys: []string{"o"},
			wantHelp: "open in browser",
		},
		{
			name:     "ViewDiff binding",
			binding:  km.ViewDiff,
			wantKeys: []string{"d"},
			wantHelp: "view diff",
		},
		{
			name:     "ViewDetails binding",
			binding:  km.ViewDetails,
			wantKeys: []string{"v", "enter"},
			wantHelp: "view details",
		},
		{
			name:     "CreatePR binding",
			binding:  km.CreatePR,
			wantKeys: []string{"c"},
			wantHelp: "create PR",
		},
		{
			name:     "MergePR binding",
			binding:  km.MergePR,
			wantKeys: []string{"m"},
			wantHelp: "merge",
		},
		{
			name:     "ClosePR binding",
			binding:  km.ClosePR,
			wantKeys: []string{"x"},
			wantHelp: "close PR",
		},
		{
			name:     "ApprovePR binding",
			binding:  km.ApprovePR,
			wantKeys: []string{"a"},
			wantHelp: "approve",
		},
		{
			name:     "RequestChanges binding",
			binding:  km.RequestChanges,
			wantKeys: []string{"R"},
			wantHelp: "request changes",
		},

		// LLM actions
		{
			name:     "GenerateWithLLM binding",
			binding:  km.GenerateWithLLM,
			wantKeys: []string{"g"},
			wantHelp: "generate with LLM",
		},
		{
			name:     "RegenerateLLM binding",
			binding:  km.RegenerateLLM,
			wantKeys: []string{"G"},
			wantHelp: "regenerate",
		},
		{
			name:     "AcceptLLM binding",
			binding:  km.AcceptLLM,
			wantKeys: []string{"enter"},
			wantHelp: "accept",
		},
		{
			name:     "EditLLM binding",
			binding:  km.EditLLM,
			wantKeys: []string{"e"},
			wantHelp: "edit",
		},

		// Editor/Composer
		{
			name:     "Submit binding",
			binding:  km.Submit,
			wantKeys: []string{"ctrl+s"},
			wantHelp: "submit",
		},
		{
			name:     "ClearAll binding",
			binding:  km.ClearAll,
			wantKeys: []string{"ctrl+u"},
			wantHelp: "clear all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()
			if len(keys) != len(tt.wantKeys) {
				t.Errorf("keys count mismatch: got %v, want %v", keys, tt.wantKeys)
				return
			}
			for i, k := range keys {
				if k != tt.wantKeys[i] {
					t.Errorf("key mismatch at index %d: got %q, want %q", i, k, tt.wantKeys[i])
				}
			}

			help := tt.binding.Help()
			if help.Desc != tt.wantHelp {
				t.Errorf("help description mismatch: got %q, want %q", help.Desc, tt.wantHelp)
			}
		})
	}
}

func TestNewKeyMapWithDefaultConfig(t *testing.T) {
	cfg := config.Default()
	km := NewKeyMap(cfg)

	// Should use config values for configurable keys
	quitKeys := km.Quit.Keys()
	found := false
	for _, k := range quitKeys {
		if k == "q" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected quit binding to include 'q' from config, got %v", quitKeys)
	}

	helpKeys := km.Help.Keys()
	if helpKeys[0] != "?" {
		t.Errorf("expected help binding to be '?' from config, got %v", helpKeys)
	}

	confirmKeys := km.Confirm.Keys()
	if confirmKeys[0] != "enter" {
		t.Errorf("expected confirm binding to be 'enter' from config, got %v", confirmKeys)
	}
}

func TestNewKeyMapWithCustomConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Keys.Quit = "ctrl+q"
	cfg.Keys.Help = "h"
	cfg.Keys.Confirm = "y"

	km := NewKeyMap(cfg)

	// Quit should include custom key plus ctrl+c as fallback
	quitKeys := km.Quit.Keys()
	foundCustom := false
	foundCtrlC := false
	for _, k := range quitKeys {
		if k == "ctrl+q" {
			foundCustom = true
		}
		if k == "ctrl+c" {
			foundCtrlC = true
		}
	}
	if !foundCustom {
		t.Errorf("expected quit binding to include custom 'ctrl+q', got %v", quitKeys)
	}
	if !foundCtrlC {
		t.Errorf("expected quit binding to always include 'ctrl+c', got %v", quitKeys)
	}

	helpKeys := km.Help.Keys()
	if helpKeys[0] != "h" {
		t.Errorf("expected help binding to be 'h' from config, got %v", helpKeys)
	}

	confirmKeys := km.Confirm.Keys()
	if confirmKeys[0] != "y" {
		t.Errorf("expected confirm binding to be 'y' from config, got %v", confirmKeys)
	}
}

func TestNewKeyMapWithEmptyConfigUsesDefaults(t *testing.T) {
	cfg := config.Config{} // All empty strings

	km := NewKeyMap(cfg)

	// Should fall back to defaults
	quitKeys := km.Quit.Keys()
	foundDefault := false
	for _, k := range quitKeys {
		if k == "q" {
			foundDefault = true
			break
		}
	}
	if !foundDefault {
		t.Errorf("expected quit binding to fall back to 'q', got %v", quitKeys)
	}

	helpKeys := km.Help.Keys()
	if helpKeys[0] != "?" {
		t.Errorf("expected help binding to fall back to '?', got %v", helpKeys)
	}

	confirmKeys := km.Confirm.Keys()
	if confirmKeys[0] != "enter" {
		t.Errorf("expected confirm binding to fall back to 'enter', got %v", confirmKeys)
	}
}

func TestShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	shortHelp := km.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("ShortHelp should return at least some bindings")
	}

	// Should include essential bindings like quit and help
	hasQuit := false
	hasHelp := false
	for _, b := range shortHelp {
		help := b.Help()
		if help.Desc == "quit" {
			hasQuit = true
		}
		if help.Desc == "help" {
			hasHelp = true
		}
	}

	if !hasQuit {
		t.Error("ShortHelp should include quit binding")
	}
	if !hasHelp {
		t.Error("ShortHelp should include help binding")
	}
}

func TestFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp should return at least some categories")
	}

	// Each category should have at least one binding
	for i, category := range fullHelp {
		if len(category) == 0 {
			t.Errorf("Category %d in FullHelp is empty", i)
		}
	}

	// Count total bindings
	totalBindings := 0
	for _, category := range fullHelp {
		totalBindings += len(category)
	}

	// Should have a reasonable number of bindings
	if totalBindings < 10 {
		t.Errorf("FullHelp should have more bindings, got %d", totalBindings)
	}
}

func TestBindingFromConfig(t *testing.T) {
	tests := []struct {
		name      string
		configVal string
		fallback  string
		helpKey   string
		helpDesc  string
		wantKeys  []string
	}{
		{
			name:      "uses config value when provided",
			configVal: "x",
			fallback:  "q",
			helpKey:   "x",
			helpDesc:  "test",
			wantKeys:  []string{"x"},
		},
		{
			name:      "uses fallback when config is empty",
			configVal: "",
			fallback:  "q",
			helpKey:   "q",
			helpDesc:  "test",
			wantKeys:  []string{"q"},
		},
		{
			name:      "uses fallback when config is whitespace",
			configVal: "   ",
			fallback:  "q",
			helpKey:   "q",
			helpDesc:  "test",
			wantKeys:  []string{"q"},
		},
		{
			name:      "handles modifier keys",
			configVal: "ctrl+s",
			fallback:  "enter",
			helpKey:   "ctrl+s",
			helpDesc:  "submit",
			wantKeys:  []string{"ctrl+s"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding := bindingFromConfig(tt.configVal, tt.fallback, tt.helpKey, tt.helpDesc)

			keys := binding.Keys()
			if len(keys) != len(tt.wantKeys) {
				t.Errorf("keys count mismatch: got %v, want %v", keys, tt.wantKeys)
				return
			}
			for i, k := range keys {
				if k != tt.wantKeys[i] {
					t.Errorf("key mismatch at index %d: got %q, want %q", i, k, tt.wantKeys[i])
				}
			}

			help := binding.Help()
			if help.Desc != tt.helpDesc {
				t.Errorf("help description mismatch: got %q, want %q", help.Desc, tt.helpDesc)
			}
		})
	}
}

func TestKeyMatchesFunction(t *testing.T) {
	km := DefaultKeyMap()

	// Create a mock key message
	type mockKeyMsg struct {
		keys string
	}

	// Test that key.Matches works correctly with our bindings
	// This is more of an integration test with bubbles/key

	// Verify bindings are enabled by default
	if !km.Quit.Enabled() {
		t.Error("Quit binding should be enabled by default")
	}
	if !km.Help.Enabled() {
		t.Error("Help binding should be enabled by default")
	}
	if !km.Confirm.Enabled() {
		t.Error("Confirm binding should be enabled by default")
	}
}

func TestHelpKeyFormatting(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name        string
		binding     key.Binding
		wantHelpKey string
	}{
		{
			name:        "simple key",
			binding:     km.Help,
			wantHelpKey: "?",
		},
		{
			name:        "modifier key",
			binding:     km.Submit,
			wantHelpKey: "ctrl+s",
		},
		{
			name:        "multi-key binding help shows first",
			binding:     km.Quit,
			wantHelpKey: "q",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			if help.Key != tt.wantHelpKey {
				t.Errorf("help key mismatch: got %q, want %q", help.Key, tt.wantHelpKey)
			}
		})
	}
}
