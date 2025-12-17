package ui

import (
	"strings"

	"bui/internal/config"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap holds all application keybindings.
// It implements the help.KeyMap interface for the bubbles/help component.
type KeyMap struct {
	// Global keys (available everywhere)
	Quit    key.Binding
	Help    key.Binding
	Confirm key.Binding
	Cancel  key.Binding

	// Navigation (arrow keys)
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// List navigation (vim style)
	NextItem key.Binding
	PrevItem key.Binding

	// Actions
	Select  key.Binding
	Back    key.Binding
	Refresh key.Binding
	Search  key.Binding
	Filter  key.Binding

	// PR-specific actions
	OpenInBrowser  key.Binding
	ViewDiff       key.Binding
	ViewDetails    key.Binding
	CreatePR       key.Binding
	MergePR        key.Binding
	ClosePR        key.Binding
	ApprovePR      key.Binding
	RequestChanges key.Binding

	// LLM actions
	GenerateWithLLM key.Binding
	RegenerateLLM   key.Binding
	AcceptLLM       key.Binding
	EditLLM         key.Binding

	// Editor/Composer
	Submit   key.Binding
	ClearAll key.Binding
}

// NewKeyMap creates a KeyMap with defaults merged with config values.
// Config values override defaults for configurable keys (quit, help, confirm).
func NewKeyMap(cfg config.Config) KeyMap {
	km := DefaultKeyMap()

	// Override configurable keys from config
	quitKey := valueOrDefault(cfg.Keys.Quit, "q")
	km.Quit = key.NewBinding(
		key.WithKeys(quitKey, "ctrl+c"),
		key.WithHelp(quitKey, "quit"),
	)

	helpKey := valueOrDefault(cfg.Keys.Help, "?")
	km.Help = key.NewBinding(
		key.WithKeys(helpKey),
		key.WithHelp(helpKey, "help"),
	)

	confirmKey := valueOrDefault(cfg.Keys.Confirm, "enter")
	km.Confirm = key.NewBinding(
		key.WithKeys(confirmKey),
		key.WithHelp(confirmKey, "confirm"),
	)

	return km
}

// DefaultKeyMap returns the default key bindings without any config overrides.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global keys
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),

		// Navigation (arrow keys)
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdown", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "go to start"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "go to end"),
		),

		// List navigation (vim style)
		NextItem: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "next"),
		),
		PrevItem: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "prev"),
		),

		// Actions
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),

		// PR-specific actions
		OpenInBrowser: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		ViewDiff: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "view diff"),
		),
		ViewDetails: key.NewBinding(
			key.WithKeys("v", "enter"),
			key.WithHelp("v", "view details"),
		),
		CreatePR: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "create PR"),
		),
		MergePR: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "merge"),
		),
		ClosePR: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "close PR"),
		),
		ApprovePR: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "approve"),
		),
		RequestChanges: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "request changes"),
		),

		// LLM actions
		GenerateWithLLM: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "generate with LLM"),
		),
		RegenerateLLM: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "regenerate"),
		),
		AcceptLLM: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "accept"),
		),
		EditLLM: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),

		// Editor/Composer
		Submit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "submit"),
		),
		ClearAll: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "clear all"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
// This implements the help.KeyMap interface.
func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.Help,
		km.Quit,
		km.Select,
		km.Back,
	}
}

// FullHelp returns keybindings for the expanded help view, organized by category.
// This implements the help.KeyMap interface.
func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Global
		{km.Help, km.Quit, km.Confirm, km.Cancel},
		// Navigation
		{km.Up, km.Down, km.NextItem, km.PrevItem, km.PageUp, km.PageDown},
		// Actions
		{km.Select, km.Back, km.Refresh, km.Search, km.Filter},
		// PR Actions
		{km.ViewDetails, km.ViewDiff, km.OpenInBrowser, km.CreatePR},
		// PR Management
		{km.ApprovePR, km.MergePR, km.ClosePR, km.RequestChanges},
		// LLM
		{km.GenerateWithLLM, km.RegenerateLLM, km.AcceptLLM, km.EditLLM},
		// Editor
		{km.Submit, km.ClearAll},
	}
}

// bindingFromConfig creates a key.Binding from a config string with fallback.
// If configVal is empty or whitespace, the fallback value is used.
func bindingFromConfig(configVal, fallback, helpKey, helpDesc string) key.Binding {
	keyVal := valueOrDefault(configVal, fallback)
	displayKey := keyVal
	if helpKey != "" {
		displayKey = helpKey
	}
	return key.NewBinding(
		key.WithKeys(keyVal),
		key.WithHelp(displayKey, helpDesc),
	)
}

// valueOrDefault returns the value if it's not empty (after trimming), otherwise returns the default.
func valueOrDefault(value, defaultValue string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultValue
	}
	return trimmed
}

// Navigation returns a subset of navigation bindings for contexts that need them.
func (km KeyMap) Navigation() []key.Binding {
	return []key.Binding{
		km.Up,
		km.Down,
		km.NextItem,
		km.PrevItem,
		km.PageUp,
		km.PageDown,
		km.Home,
		km.End,
	}
}

// PRActions returns PR-specific action bindings.
func (km KeyMap) PRActions() []key.Binding {
	return []key.Binding{
		km.ViewDetails,
		km.ViewDiff,
		km.OpenInBrowser,
		km.CreatePR,
		km.ApprovePR,
		km.MergePR,
		km.ClosePR,
		km.RequestChanges,
	}
}

// LLMActions returns LLM-specific action bindings.
func (km KeyMap) LLMActions() []key.Binding {
	return []key.Binding{
		km.GenerateWithLLM,
		km.RegenerateLLM,
		km.AcceptLLM,
		km.EditLLM,
	}
}

// EditorActions returns editor/composer action bindings.
func (km KeyMap) EditorActions() []key.Binding {
	return []key.Binding{
		km.Submit,
		km.ClearAll,
	}
}
