package help

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model. It handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// Window size change
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	// Keyboard input
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Pass remaining messages to viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Close help screen
	case key.Matches(msg, m.keymap.Cancel):
		return m, closeHelp()

	case key.Matches(msg, m.keymap.Quit):
		// q closes help, doesn't quit app
		return m, closeHelp()

	case key.Matches(msg, m.keymap.Help):
		// ? toggles help off when already in help
		return m, closeHelp()

	// Scroll down (j or down arrow)
	case key.Matches(msg, m.keymap.NextItem), key.Matches(msg, m.keymap.Down):
		m.viewport.LineDown(1)
		return m, nil

	// Scroll up (k or up arrow)
	case key.Matches(msg, m.keymap.PrevItem), key.Matches(msg, m.keymap.Up):
		m.viewport.LineUp(1)
		return m, nil

	// Page down
	case key.Matches(msg, m.keymap.PageDown):
		m.viewport.ViewDown()
		return m, nil

	// Page up
	case key.Matches(msg, m.keymap.PageUp):
		m.viewport.ViewUp()
		return m, nil

	// Go to top (Home or g)
	case key.Matches(msg, m.keymap.Home):
		m.viewport.GotoTop()
		return m, nil

	case msg.String() == "g":
		m.viewport.GotoTop()
		return m, nil

	// Go to bottom (End or G)
	case key.Matches(msg, m.keymap.End):
		m.viewport.GotoBottom()
		return m, nil

	case msg.String() == "G":
		m.viewport.GotoBottom()
		return m, nil
	}

	// Pass to viewport for any other scrolling behavior
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}
