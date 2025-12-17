package review

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
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
	case GenerateCommentStartMsg:
		return m.handleGenerateStart()

	case GenerateCommentTokenMsg:
		return m.handleGenerateToken(msg)

	case GenerateCommentDoneMsg:
		return m.handleGenerateDone()

	case GenerateCommentErrorMsg:
		return m.handleGenerateError(msg)

	// Submission messages
	case SubmitReviewMsg:
		return m.handleSubmitReview()

	case ReviewSubmittedMsg:
		return m.handleReviewSubmitted()

	case ReviewSubmitErrorMsg:
		return m.handleReviewSubmitError(msg)

	// Spinner tick
	case spinner.TickMsg:
		if m.state == StateSubmitting || m.isGenerating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	// Keyboard input
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Pass remaining messages to textarea when in compose state
	if m.state == StateCompose && !m.isGenerating {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		// Update comment from textarea
		m.comment = m.textarea.Value()
		cmds = append(cmds, cmd)
	}

	// Pass remaining messages to viewport when in preview state
	if m.state == StatePreview {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
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
func (m Model) handleGenerateToken(msg GenerateCommentTokenMsg) (tea.Model, tea.Cmd) {
	m.generatedComment += msg.Token
	return m, m.readNextToken()
}

// handleGenerateDone completes the LLM generation.
func (m Model) handleGenerateDone() (tea.Model, tea.Cmd) {
	m.isGenerating = false
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}

	// Copy generated content to textarea for editing
	if m.generatedComment != "" {
		m.textarea.SetValue(m.generatedComment)
		m.comment = m.generatedComment
	}

	return m, nil
}

// handleGenerateError processes LLM errors.
func (m Model) handleGenerateError(msg GenerateCommentErrorMsg) (tea.Model, tea.Cmd) {
	m.isGenerating = false
	m.err = msg.Err
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}
	return m, nil
}

// =============================================================================
// Submission Handlers
// =============================================================================

// handleSubmitReview starts the review submission.
func (m Model) handleSubmitReview() (tea.Model, tea.Cmd) {
	m.state = StateSubmitting
	m.spinner = m.spinner.WithMessage("Submitting review...")
	return m, tea.Batch(
		m.spinner.Init(),
		m.submitReview(),
	)
}

// handleReviewSubmitted processes successful submission.
func (m Model) handleReviewSubmitted() (tea.Model, tea.Cmd) {
	m.state = StateComplete
	return m, nil
}

// handleReviewSubmitError processes submission errors.
func (m Model) handleReviewSubmitError(msg ReviewSubmitErrorMsg) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = msg.Err
	return m, nil
}

// =============================================================================
// Key Handlers
// =============================================================================

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle state-specific key bindings
	switch m.state {
	case StateTypeSelect:
		return m.handleTypeSelectKeys(msg)
	case StateCompose:
		return m.handleComposeKeys(msg)
	case StatePreview:
		return m.handlePreviewKeys(msg)
	case StateSubmitting:
		return m.handleSubmittingKeys(msg)
	case StateComplete:
		return m.handleCompleteKeys(msg)
	case StateError:
		return m.handleErrorKeys(msg)
	}

	return m, nil
}

// handleTypeSelectKeys handles keys during type selection state.
func (m Model) handleTypeSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	// 1 or a - Approve
	case "1", "a":
		m.reviewType = ReviewTypeApprove
		m.state = StateCompose
		m.textarea.Focus()
		m.textarea.Placeholder = "Add a comment (optional for approve)..."
		return m, textarea.Blink

	// 2 or c - Request Changes
	case "2", "c":
		m.reviewType = ReviewTypeRequestChanges
		m.state = StateCompose
		m.textarea.Focus()
		m.textarea.Placeholder = "Describe the changes needed (required)..."
		return m, textarea.Blink

	// 3 or m - Comment
	case "3", "m":
		m.reviewType = ReviewTypeComment
		m.state = StateCompose
		m.textarea.Focus()
		m.textarea.Placeholder = "Enter your comment (required)..."
		return m, textarea.Blink

	// Escape - cancel
	case "esc":
		return m, backToPRDetail(m.prNumber)

	// Quit
	case "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// handleComposeKeys handles keys during compose state.
func (m Model) handleComposeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// If generating, only allow cancel
	if m.isGenerating {
		switch {
		case msg.Type == tea.KeyCtrlX || msg.Type == tea.KeyEscape:
			m.cancelGeneration()
			return m, nil
		case key.Matches(msg, m.keymap.Quit):
			if m.llmCancel != nil {
				m.llmCancel()
			}
			return m, tea.Quit
		}
		return m, nil
	}

	switch {
	// Escape - go back to type select
	case msg.Type == tea.KeyEscape:
		m.state = StateTypeSelect
		return m, nil

	// Ctrl+G - generate with LLM
	case msg.Type == tea.KeyCtrlG:
		if m.llmProvider != nil {
			m.spinner = m.spinner.WithMessage("Generating comment...")
			cmd := m.startGeneration()
			return m, cmd
		}
		return m, nil

	// Ctrl+S - proceed to preview
	case msg.Type == tea.KeyCtrlS:
		// Update comment from textarea
		m.comment = m.textarea.Value()

		// Check if we can proceed
		if m.canProceedToPreview() {
			m.state = StatePreview
			// Update viewport content
			m.viewport.SetContent(m.buildPreviewContent())
			return m, nil
		}
		// Stay in compose if validation fails
		return m, nil

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	default:
		// Pass to textarea
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		m.comment = m.textarea.Value()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handlePreviewKeys handles keys during preview state.
func (m Model) handlePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Enter - submit review
	case msg.Type == tea.KeyEnter:
		m.state = StateSubmitting
		m.spinner = m.spinner.WithMessage("Submitting review...")
		return m, tea.Batch(
			m.spinner.Init(),
			m.submitReview(),
		)

	// e - edit (go back to compose)
	case msg.String() == "e":
		m.state = StateCompose
		m.textarea.Focus()
		return m, textarea.Blink

	// Escape - go back to compose
	case msg.Type == tea.KeyEscape:
		m.state = StateCompose
		m.textarea.Focus()
		return m, textarea.Blink

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	}

	// Pass to viewport for scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// handleSubmittingKeys handles keys during submitting state.
func (m Model) handleSubmittingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only allow quit during submission
	if key.Matches(msg, m.keymap.Quit) {
		return m, tea.Quit
	}
	return m, nil
}

// handleCompleteKeys handles keys during complete state.
func (m Model) handleCompleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Enter or Escape - go back to PR detail
	case msg.Type == tea.KeyEnter || msg.Type == tea.KeyEscape:
		return m, backToPRDetail(m.prNumber)

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	}

	return m, nil
}

// handleErrorKeys handles keys during error state.
func (m Model) handleErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// r - retry (go back to preview)
	case msg.String() == "r":
		m.state = StatePreview
		m.err = nil
		return m, nil

	// Escape - go back to compose
	case msg.Type == tea.KeyEscape:
		m.state = StateCompose
		m.err = nil
		m.textarea.Focus()
		return m, textarea.Blink

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	}

	return m, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// buildPreviewContent creates the content for the preview viewport.
func (m Model) buildPreviewContent() string {
	var content string
	content += "Review Type: " + m.reviewType.DisplayName() + "\n\n"

	if m.comment != "" {
		content += "Comment:\n"
		content += m.comment
	} else {
		content += "(No comment)"
	}

	return content
}
