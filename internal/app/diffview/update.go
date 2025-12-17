package diffview

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// update handles all incoming messages and returns the updated model and any commands.
func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		// Handle key input based on state
		if m.state == StateLoading {
			// Only allow quit during loading
			if key.Matches(msg, m.keymap.Quit) {
				return m, tea.Quit
			}
			return m, nil
		}

		if m.state == StateError {
			// Allow quit or back on error
			switch {
			case key.Matches(msg, m.keymap.Quit):
				return m, tea.Quit
			case key.Matches(msg, m.keymap.Back):
				return m, backToPRDetail(m.prNumber)
			}
			return m, nil
		}

		// Ready state - full key handling
		return m.handleKeyMsg(msg)

	case DiffLoadedMsg:
		return m.handleDiffLoaded(msg)

	case DiffErrorMsg:
		m.state = StateError
		m.err = msg.Err
		return m, nil

	case FileSelectedMsg:
		m.selectFileByIndex(msg.Index)
		return m, nil

	case spinner.TickMsg:
		if m.state == StateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Pass other messages to sub-components
	return m.updateComponents(msg)
}

// handleKeyMsg processes keyboard input in the ready state.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Global keys
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keymap.Back):
		return m, backToPRDetail(m.prNumber)

	// Tab to switch focus
	case msg.String() == "tab":
		m.toggleFocus()
		return m, nil

	// Toggle line numbers
	case msg.String() == "n":
		m.toggleLineNumbers()
		return m, nil

	// Search in diff (when focused on diff)
	case msg.String() == "/" && m.focus == FocusDiff:
		// TODO: Implement search
		return m, nil

	// Space to toggle hunk selection (when focused on diff)
	case msg.String() == " " && m.focus == FocusDiff:
		// TODO: Implement hunk selection mode
		return m, nil
	}

	// Handle focus-specific navigation
	if m.focus == FocusFileTree {
		return m.handleFileTreeKeys(msg)
	}
	return m.handleDiffKeys(msg)
}

// handleFileTreeKeys handles keyboard input when file tree has focus.
func (m Model) handleFileTreeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Enter to select file
	case key.Matches(msg, m.keymap.Select):
		if item, ok := m.fileList.SelectedItem().(FileItem); ok {
			m.selectFileByIndex(item.Index)
			m.focus = FocusDiff // Switch to diff view after selection
		}
		return m, nil

	// Navigation
	case key.Matches(msg, m.keymap.NextItem) || key.Matches(msg, m.keymap.Down):
		m.fileList.CursorDown()
		// Also update selected file preview
		if item, ok := m.fileList.SelectedItem().(FileItem); ok {
			m.selectFileByIndex(item.Index)
		}
		return m, nil

	case key.Matches(msg, m.keymap.PrevItem) || key.Matches(msg, m.keymap.Up):
		m.fileList.CursorUp()
		// Also update selected file preview
		if item, ok := m.fileList.SelectedItem().(FileItem); ok {
			m.selectFileByIndex(item.Index)
		}
		return m, nil

	case key.Matches(msg, m.keymap.Home):
		m.fileList.Select(0)
		if item, ok := m.fileList.SelectedItem().(FileItem); ok {
			m.selectFileByIndex(item.Index)
		}
		return m, nil

	case key.Matches(msg, m.keymap.End):
		m.fileList.Select(len(m.files) - 1)
		if item, ok := m.fileList.SelectedItem().(FileItem); ok {
			m.selectFileByIndex(item.Index)
		}
		return m, nil
	}

	// Pass to list for other handling (like filtering)
	var cmd tea.Cmd
	m.fileList, cmd = m.fileList.Update(msg)
	return m, cmd
}

// handleDiffKeys handles keyboard input when diff panel has focus.
func (m Model) handleDiffKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Scroll navigation
	case key.Matches(msg, m.keymap.NextItem) || key.Matches(msg, m.keymap.Down):
		m.diffViewport.LineDown(1)
		return m, nil

	case key.Matches(msg, m.keymap.PrevItem) || key.Matches(msg, m.keymap.Up):
		m.diffViewport.LineUp(1)
		return m, nil

	case key.Matches(msg, m.keymap.PageDown):
		m.diffViewport.HalfViewDown()
		return m, nil

	case key.Matches(msg, m.keymap.PageUp):
		m.diffViewport.HalfViewUp()
		return m, nil

	case key.Matches(msg, m.keymap.Home):
		m.diffViewport.GotoTop()
		return m, nil

	case key.Matches(msg, m.keymap.End):
		m.diffViewport.GotoBottom()
		return m, nil

	// Left arrow to go back to file tree
	case key.Matches(msg, m.keymap.Left):
		m.focus = FocusFileTree
		return m, nil

	// ] and [ to navigate between files while in diff view
	case msg.String() == "]":
		nextIdx := m.selectedFile + 1
		if nextIdx < len(m.files) {
			m.selectFileByIndex(nextIdx)
			m.fileList.Select(nextIdx)
		}
		return m, nil

	case msg.String() == "[":
		prevIdx := m.selectedFile - 1
		if prevIdx >= 0 {
			m.selectFileByIndex(prevIdx)
			m.fileList.Select(prevIdx)
		}
		return m, nil
	}

	// Pass to viewport for other handling
	var cmd tea.Cmd
	m.diffViewport, cmd = m.diffViewport.Update(msg)
	return m, cmd
}

// handleDiffLoaded processes a successfully loaded diff.
func (m Model) handleDiffLoaded(msg DiffLoadedMsg) (tea.Model, tea.Cmd) {
	m.rawDiff = msg.Diff
	m.files = ParseDiff(msg.Diff)
	m.state = StateReady

	// Update the file list with parsed files
	m.updateFileList()

	// Select the first file if available
	if len(m.files) > 0 {
		m.selectFileByIndex(0)
	}

	return m, nil
}

// updateComponents passes messages to sub-components and collects commands.
func (m Model) updateComponents(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update file list
	var fileListCmd tea.Cmd
	m.fileList, fileListCmd = m.fileList.Update(msg)
	if fileListCmd != nil {
		cmds = append(cmds, fileListCmd)
	}

	// Update viewport
	var viewportCmd tea.Cmd
	m.diffViewport, viewportCmd = m.diffViewport.Update(msg)
	if viewportCmd != nil {
		cmds = append(cmds, viewportCmd)
	}

	// Handle list item selection
	if msg, ok := msg.(list.FilterMatchesMsg); ok && len(msg) > 0 {
		// After filtering, select the first matched item
		if item, ok := m.fileList.SelectedItem().(FileItem); ok {
			m.selectFileByIndex(item.Index)
		}
	}

	return m, tea.Batch(cmds...)
}
