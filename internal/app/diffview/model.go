package diffview

import (
	"bui/internal/config"
	"bui/internal/gh"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// defaultFileListWidth is the default width percentage for the file list panel.
	defaultFileListWidthPercent = 25

	// minFileListWidth is the minimum width for the file list panel.
	minFileListWidth = 20

	// maxFileListWidth is the maximum width for the file list panel.
	maxFileListWidth = 60

	// footerHeight is the height reserved for the footer/help bar.
	footerHeight = 2

	// headerHeight is the height reserved for the header.
	headerHeight = 1

	// borderWidth accounts for borders between panels.
	borderWidth = 1
)

// =============================================================================
// State
// =============================================================================

// State represents the current state of the diff view screen.
type State int

const (
	// StateLoading indicates data is being loaded.
	StateLoading State = iota
	// StateReady indicates data is loaded and ready for interaction.
	StateReady
	// StateError indicates an error occurred.
	StateError
)

// String returns a string representation of the state.
func (s State) String() string {
	switch s {
	case StateLoading:
		return "loading"
	case StateReady:
		return "ready"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// =============================================================================
// Focus
// =============================================================================

// Focus represents which panel has focus.
type Focus int

const (
	// FocusFileTree indicates the file tree panel has focus.
	FocusFileTree Focus = iota
	// FocusDiff indicates the diff content panel has focus.
	FocusDiff
)

// String returns a string representation of the focus.
func (f Focus) String() string {
	switch f {
	case FocusFileTree:
		return "file_tree"
	case FocusDiff:
		return "diff"
	default:
		return "unknown"
	}
}

// =============================================================================
// Model
// =============================================================================

// Model is the Bubble Tea model for the diff view screen.
type Model struct {
	// Dependencies
	cfg      config.Config
	ghClient *gh.Client

	// Data
	prNumber int
	rawDiff  string
	files    []FileDiff

	// UI state
	state         State
	focus         Focus
	fileList      list.Model
	diffViewport  viewport.Model
	selectedFile  int
	showLineNums  bool
	hunkSelection map[string]map[int]bool // map[filepath]map[hunkIdx]selected
	err           error

	// Spinner for loading state
	spinner ui.Spinner

	// Styling
	palette ui.Palette
	styles  ui.Styles
	keymap  ui.KeyMap

	// Dimensions
	width, height int
	fileListWidth int // Actual width of file tree panel
}

// New creates a new diff view model.
func New(cfg config.Config, prNumber int) Model {
	return NewWithClient(cfg, prNumber, gh.New())
}

// NewWithClient creates a new diff view model with a custom GitHub client.
// This is useful for testing.
func NewWithClient(cfg config.Config, prNumber int, client *gh.Client) Model {
	palette := ui.NewPalette(cfg)
	styles := ui.NewStyles(palette)
	keymap := ui.NewKeyMap(cfg)

	// Initialize empty file list (will be populated when diff loads)
	fileList := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	fileList.SetShowHelp(false)
	fileList.SetShowStatusBar(false)
	fileList.SetFilteringEnabled(false)
	fileList.DisableQuitKeybindings()

	// Initialize viewport (will be resized on first WindowSizeMsg)
	vp := viewport.New(80, 20)
	vp.Style = styles.App

	// Initialize spinner
	spinner := ui.NewSpinner(palette, styles).WithMessage("Loading diff...")

	return Model{
		cfg:           cfg,
		ghClient:      client,
		prNumber:      prNumber,
		state:         StateLoading,
		focus:         FocusFileTree,
		fileList:      fileList,
		diffViewport:  vp,
		selectedFile:  0,
		showLineNums:  cfg.UI.ShowLineNumbers,
		hunkSelection: make(map[string]map[int]bool),
		spinner:       spinner,
		palette:       palette,
		styles:        styles,
		keymap:        keymap,
		fileListWidth: minFileListWidth,
	}
}

// Init implements tea.Model. It returns the initial commands to load data.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadDiff(m.ghClient, m.prNumber),
		m.spinner.Init(),
	)
}

// Update implements tea.Model. It handles incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.update(msg)
}

// View implements tea.Model. It renders the view.
func (m Model) View() string {
	return m.view()
}

// =============================================================================
// Accessors
// =============================================================================

// SetSize updates the model dimensions and recalculates panel sizes.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate file list width (percentage-based)
	m.fileListWidth = (width * defaultFileListWidthPercent) / 100
	if m.fileListWidth < minFileListWidth {
		m.fileListWidth = minFileListWidth
	}
	if m.fileListWidth > maxFileListWidth {
		m.fileListWidth = maxFileListWidth
	}

	// Calculate available height for content
	contentHeight := height - footerHeight - headerHeight
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Update file list dimensions
	m.fileList.SetSize(m.fileListWidth-2, contentHeight)

	// Update viewport dimensions
	viewportWidth := width - m.fileListWidth - borderWidth - 2
	if viewportWidth < 20 {
		viewportWidth = 20
	}
	m.diffViewport.Width = viewportWidth
	m.diffViewport.Height = contentHeight
}

// CurrentFile returns the currently selected file diff, or nil if none.
func (m Model) CurrentFile() *FileDiff {
	if m.selectedFile < 0 || m.selectedFile >= len(m.files) {
		return nil
	}
	return &m.files[m.selectedFile]
}

// PRNumber returns the PR number being viewed.
func (m Model) PRNumber() int {
	return m.prNumber
}

// State returns the current state.
func (m Model) State() State {
	return m.state
}

// Focus returns the current focus.
func (m Model) Focus() Focus {
	return m.focus
}

// Files returns the parsed file diffs.
func (m Model) Files() []FileDiff {
	return m.files
}

// SelectedFileIndex returns the index of the currently selected file.
func (m Model) SelectedFileIndex() int {
	return m.selectedFile
}

// ShowLineNumbers returns whether line numbers are enabled.
func (m Model) ShowLineNumbers() bool {
	return m.showLineNums
}

// Error returns the current error, if any.
func (m Model) Error() error {
	return m.err
}

// IsHunkSelected returns whether a specific hunk is selected.
func (m Model) IsHunkSelected(filePath string, hunkIdx int) bool {
	if fileHunks, ok := m.hunkSelection[filePath]; ok {
		return fileHunks[hunkIdx]
	}
	return false
}

// SelectedHunks returns all currently selected hunks.
func (m Model) SelectedHunks() []SelectHunkMsg {
	var selected []SelectHunkMsg
	for filePath, hunks := range m.hunkSelection {
		for hunkIdx, isSelected := range hunks {
			if isSelected {
				// Find the file and hunk content
				for _, f := range m.files {
					if f.DisplayPath() == filePath && hunkIdx < len(f.Hunks) {
						selected = append(selected, SelectHunkMsg{
							File:     filePath,
							HunkIdx:  hunkIdx,
							Content:  f.Hunks[hunkIdx].Header,
							Selected: true,
						})
						break
					}
				}
			}
		}
	}
	return selected
}

// =============================================================================
// Internal Helpers
// =============================================================================

// toggleHunkSelection toggles the selection state of a hunk.
func (m *Model) toggleHunkSelection(filePath string, hunkIdx int) {
	if m.hunkSelection[filePath] == nil {
		m.hunkSelection[filePath] = make(map[int]bool)
	}
	m.hunkSelection[filePath][hunkIdx] = !m.hunkSelection[filePath][hunkIdx]
}

// updateFileList updates the file list from the parsed files.
func (m *Model) updateFileList() {
	if len(m.files) == 0 {
		return
	}
	m.fileList = NewFileList(m.files, m.palette, m.styles)
	m.fileList.SetSize(m.fileListWidth-2, m.height-footerHeight-headerHeight)
}

// updateDiffContent updates the viewport content for the selected file.
func (m *Model) updateDiffContent() {
	content := m.renderCurrentFileDiff()
	m.diffViewport.SetContent(content)
	m.diffViewport.GotoTop()
}

// selectFileByIndex selects a file by its index and updates the view.
func (m *Model) selectFileByIndex(idx int) {
	if idx < 0 || idx >= len(m.files) {
		return
	}
	m.selectedFile = idx
	m.fileList.Select(idx)
	m.updateDiffContent()
}

// toggleFocus switches focus between file tree and diff panel.
func (m *Model) toggleFocus() {
	if m.focus == FocusFileTree {
		m.focus = FocusDiff
	} else {
		m.focus = FocusFileTree
	}
}

// toggleLineNumbers toggles the line number display.
func (m *Model) toggleLineNumbers() {
	m.showLineNums = !m.showLineNums
	m.updateDiffContent()
}
