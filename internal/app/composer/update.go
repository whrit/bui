package composer

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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

	// LLM generation messages
	case GenerateStartMsg:
		return m.handleGenerateStart()

	case GenerateTokenMsg:
		return m.handleGenerateToken(msg)

	case GenerateDoneMsg:
		return m.handleGenerateDone()

	case GenerateErrorMsg:
		return m.handleGenerateError(msg)

	// Spinner tick
	case spinner.TickMsg:
		if m.state == StateGenerating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	// Keyboard input
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Pass remaining messages to textarea when editing
	if m.state == StateEditing || m.state == StateReady {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// =============================================================================
// LLM Handlers
// =============================================================================

// handleGenerateStart begins the LLM generation.
func (m Model) handleGenerateStart() (tea.Model, tea.Cmd) {
	cmd := m.startGeneration()
	return m, cmd
}

// handleGenerateToken processes an incoming LLM token.
func (m Model) handleGenerateToken(msg GenerateTokenMsg) (tea.Model, tea.Cmd) {
	m.streamedText += msg.Token
	return m, m.readNextToken()
}

// handleGenerateDone completes the LLM generation.
func (m Model) handleGenerateDone() (tea.Model, tea.Cmd) {
	m.state = StateReady
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}

	// Copy generated content to textarea for editing
	if m.streamedText != "" {
		m.textarea.SetValue(m.streamedText)
	}

	return m, nil
}

// handleGenerateError processes LLM errors.
func (m Model) handleGenerateError(msg GenerateErrorMsg) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = msg.Err
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}
	return m, nil
}

// =============================================================================
// Key Handlers
// =============================================================================

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle state-specific key bindings
	switch m.state {
	case StateEditing:
		return m.handleEditingKeys(msg)
	case StateGenerating:
		return m.handleGeneratingKeys(msg)
	case StateReady:
		return m.handleReadyKeys(msg)
	case StateError:
		return m.handleErrorKeys(msg)
	}

	return m, nil
}

// handleEditingKeys handles keys during editing state.
func (m Model) handleEditingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	// Escape - cancel and go back
	case key.Matches(msg, m.keymap.Cancel) || msg.Type == tea.KeyEscape:
		if m.onCancel != nil {
			return m, m.onCancel()
		}
		return m, cancelComposer()

	// Ctrl+G - generate with LLM
	case msg.Type == tea.KeyCtrlG:
		if m.llmProvider != nil {
			cmd := m.startGeneration()
			return m, cmd
		}
		return m, nil

	// Ctrl+S - submit content
	case msg.Type == tea.KeyCtrlS:
		content := m.GetContent()
		if m.onSubmit != nil {
			return m, m.onSubmit(content)
		}
		return m, acceptContent(content)

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	default:
		// Pass to textarea
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleGeneratingKeys handles keys during LLM generation.
func (m Model) handleGeneratingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Ctrl+X - cancel generation
	case msg.Type == tea.KeyCtrlX:
		m.cancelGeneration()
		return m, nil

	// Escape - cancel generation
	case msg.Type == tea.KeyEscape:
		m.cancelGeneration()
		return m, nil

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		if m.llmCancel != nil {
			m.llmCancel()
		}
		return m, tea.Quit
	}

	return m, nil
}

// handleReadyKeys handles keys when generation is complete (ready state).
func (m Model) handleReadyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	// Escape - cancel and go back
	case key.Matches(msg, m.keymap.Cancel) || msg.Type == tea.KeyEscape:
		if m.onCancel != nil {
			return m, m.onCancel()
		}
		return m, cancelComposer()

	// Ctrl+S or Enter - accept content
	case msg.Type == tea.KeyCtrlS:
		content := m.GetContent()
		if m.onSubmit != nil {
			return m, m.onSubmit(content)
		}
		return m, acceptContent(content)

	// Ctrl+G or G - regenerate
	case msg.Type == tea.KeyCtrlG:
		if m.llmProvider != nil {
			cmd := m.startGeneration()
			return m, cmd
		}
		return m, nil

	// e - edit the generated content (switch to editing mode)
	case msg.String() == "e":
		m.state = StateEditing
		m.textarea.Focus()
		return m, nil

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	default:
		// Pass to textarea for editing
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleErrorKeys handles keys during error state.
func (m Model) handleErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Escape - go back
	case key.Matches(msg, m.keymap.Cancel) || msg.Type == tea.KeyEscape:
		if m.onCancel != nil {
			return m, m.onCancel()
		}
		return m, cancelComposer()

	// r - retry generation
	case msg.String() == "r":
		m.state = StateEditing
		m.err = nil
		return m, nil

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	}

	return m, nil
}
