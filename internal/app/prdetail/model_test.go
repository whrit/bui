package prdetail

import (
	"context"
	"errors"
	"testing"
	"time"

	"bui/internal/config"
	"bui/internal/gh"
	"bui/internal/llm"
	"bui/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Test Fixtures
// =============================================================================

// testConfig returns a default test configuration.
func testConfig() config.Config {
	return config.Config{
		UI: config.UIConfig{
			Theme:  "dark",
			Accent: "indigo",
		},
		Keys: config.KeysConfig{
			Quit:    "q",
			Help:    "?",
			Confirm: "enter",
		},
		Behavior: config.BehaviorConfig{
			ConfirmMerge: true,
			AutoRefresh:  false,
		},
	}
}

// testPR returns a test PR fixture.
func testPR() *gh.PR {
	return &gh.PR{
		Number:     123,
		Title:      "Add new feature",
		Body:       "This PR adds a new feature.\n\n## Changes\n- Added feature X\n- Updated tests",
		State:      "OPEN",
		Author:     gh.User{Login: "testuser"},
		HeadBranch: "feature/new-feature",
		BaseBranch: "main",
		URL:        "https://github.com/test/repo/pull/123",
		CreatedAt:  time.Now().Add(-24 * time.Hour),
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
		IsDraft:    false,
		Mergeable:  "MERGEABLE",
		Additions:  150,
		Deletions:  30,
		Labels: []gh.Label{
			{Name: "enhancement", Color: "84b6eb"},
			{Name: "review-needed", Color: "fbca04"},
		},
		Reviewers: []gh.Reviewer{
			{Login: "reviewer1", State: "PENDING"},
			{Login: "reviewer2", State: "APPROVED"},
		},
	}
}

// testChecks returns test CI check fixtures.
func testChecks() []gh.Check {
	return []gh.Check{
		{Name: "build", Status: "completed", Conclusion: "success", URL: "https://example.com/build"},
		{Name: "lint", Status: "completed", Conclusion: "success", URL: "https://example.com/lint"},
		{Name: "test", Status: "completed", Conclusion: "failure", URL: "https://example.com/test"},
		{Name: "deploy-preview", Status: "in_progress", Conclusion: "", URL: "https://example.com/deploy"},
	}
}

// testReviews returns test review fixtures.
func testReviews() []gh.Review {
	return []gh.Review{
		{
			ID:        1,
			Author:    gh.User{Login: "reviewer1"},
			State:     "APPROVED",
			Body:      "Looks good to me!",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        2,
			Author:    gh.User{Login: "reviewer2"},
			State:     "CHANGES_REQUESTED",
			Body:      "Please fix the typo on line 42.",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}
}

// =============================================================================
// Mock LLM Provider
// =============================================================================

// mockLLMProvider implements llm.Provider for testing.
type mockLLMProvider struct {
	name      string
	tokens    []string
	streamErr error
}

func newMockLLMProvider(tokens []string, err error) *mockLLMProvider {
	return &mockLLMProvider{
		name:      "mock",
		tokens:    tokens,
		streamErr: err,
	}
}

func (p *mockLLMProvider) Name() string {
	return p.name
}

func (p *mockLLMProvider) Stream(ctx context.Context, prompt llm.Prompt) (<-chan llm.Token, <-chan error) {
	tokenCh := make(chan llm.Token, len(p.tokens))
	errCh := make(chan error, 1)

	go func() {
		defer close(tokenCh)
		defer close(errCh)

		if p.streamErr != nil {
			errCh <- p.streamErr
			return
		}

		for _, t := range p.tokens {
			select {
			case <-ctx.Done():
				return
			case tokenCh <- llm.Token{Text: t}:
			}
		}
	}()

	return tokenCh, errCh
}

// =============================================================================
// Model Tests
// =============================================================================

func TestNew(t *testing.T) {
	t.Run("creates model with correct initial state", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)

		if m.prNumber != 123 {
			t.Errorf("expected prNumber=123, got %d", m.prNumber)
		}

		if m.state != StateLoading {
			t.Errorf("expected state=StateLoading, got %s", m.state)
		}

		if m.activeSection != SectionInfo {
			t.Errorf("expected activeSection=SectionInfo, got %s", m.activeSection)
		}

		if m.pr != nil {
			t.Error("expected pr to be nil initially")
		}
	})

	t.Run("initializes with default merge method", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)

		if m.mergeMethod != gh.MergeMethodSquash {
			t.Errorf("expected mergeMethod=squash, got %s", m.mergeMethod)
		}
	})
}

func TestPRNumber(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 456)

	if m.PRNumber() != 456 {
		t.Errorf("expected PRNumber()=456, got %d", m.PRNumber())
	}
}

func TestState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	if m.State() != StateLoading {
		t.Errorf("expected State()=StateLoading, got %s", m.State())
	}
}

func TestActiveSection(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	if m.ActiveSection() != SectionInfo {
		t.Errorf("expected ActiveSection()=SectionInfo, got %s", m.ActiveSection())
	}
}

func TestSectionNavigation(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	// Simulate data loaded
	m.pr = testPR()
	m.prLoaded = true
	m.checksLoaded = true
	m.reviewsLoaded = true
	m.diffLoaded = true
	m.state = StateReady

	t.Run("nextSection cycles forward", func(t *testing.T) {
		m.activeSection = SectionInfo
		m.nextSection()
		if m.activeSection != SectionDescription {
			t.Errorf("expected SectionDescription, got %s", m.activeSection)
		}

		m.nextSection()
		if m.activeSection != SectionChecks {
			t.Errorf("expected SectionChecks, got %s", m.activeSection)
		}

		m.nextSection()
		if m.activeSection != SectionReviews {
			t.Errorf("expected SectionReviews, got %s", m.activeSection)
		}

		m.nextSection()
		if m.activeSection != SectionInfo {
			t.Errorf("expected SectionInfo (wrapped), got %s", m.activeSection)
		}
	})

	t.Run("prevSection cycles backward", func(t *testing.T) {
		m.activeSection = SectionInfo
		m.prevSection()
		if m.activeSection != SectionReviews {
			t.Errorf("expected SectionReviews (wrapped), got %s", m.activeSection)
		}

		m.prevSection()
		if m.activeSection != SectionChecks {
			t.Errorf("expected SectionChecks, got %s", m.activeSection)
		}
	})
}

func TestCanMerge(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name     string
		pr       *gh.PR
		expected bool
	}{
		{
			name:     "nil PR",
			pr:       nil,
			expected: false,
		},
		{
			name: "open and mergeable",
			pr: &gh.PR{
				State:     "OPEN",
				Mergeable: "MERGEABLE",
			},
			expected: true,
		},
		{
			name: "open but conflicting",
			pr: &gh.PR{
				State:     "OPEN",
				Mergeable: "CONFLICTING",
			},
			expected: false,
		},
		{
			name: "closed",
			pr: &gh.PR{
				State:     "CLOSED",
				Mergeable: "MERGEABLE",
			},
			expected: false,
		},
		{
			name: "merged",
			pr: &gh.PR{
				State:     "MERGED",
				Mergeable: "MERGEABLE",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(cfg, 123)
			m.pr = tt.pr

			if m.canMerge() != tt.expected {
				t.Errorf("expected canMerge()=%v, got %v", tt.expected, m.canMerge())
			}
		})
	}
}

func TestCanClose(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name     string
		pr       *gh.PR
		expected bool
	}{
		{
			name:     "nil PR",
			pr:       nil,
			expected: false,
		},
		{
			name: "open PR",
			pr: &gh.PR{
				State: "OPEN",
			},
			expected: true,
		},
		{
			name: "closed PR",
			pr: &gh.PR{
				State: "CLOSED",
			},
			expected: false,
		},
		{
			name: "merged PR",
			pr: &gh.PR{
				State: "MERGED",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(cfg, 123)
			m.pr = tt.pr

			if m.canClose() != tt.expected {
				t.Errorf("expected canClose()=%v, got %v", tt.expected, m.canClose())
			}
		})
	}
}

func TestIsLoading(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	if !m.IsLoading() {
		t.Error("expected IsLoading()=true initially")
	}

	m.prLoaded = true
	if !m.IsLoading() {
		t.Error("expected IsLoading()=true with partial load")
	}

	m.checksLoaded = true
	m.reviewsLoaded = true
	m.diffLoaded = true

	if m.IsLoading() {
		t.Error("expected IsLoading()=false when all loaded")
	}
}

// =============================================================================
// Update Tests
// =============================================================================

func TestUpdate_PRLoaded(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	// Pre-load other data
	m.checksLoaded = true
	m.reviewsLoaded = true
	m.diffLoaded = true

	pr := testPR()
	msg := PRLoadedMsg{PR: pr}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.pr == nil {
		t.Fatal("expected pr to be set")
	}

	if m.pr.Number != 123 {
		t.Errorf("expected pr.Number=123, got %d", m.pr.Number)
	}

	if m.state != StateReady {
		t.Errorf("expected state=StateReady, got %s", m.state)
	}
}

func TestUpdate_ChecksLoaded(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	checks := testChecks()
	msg := ChecksLoadedMsg{Checks: checks}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if len(m.checks) != 4 {
		t.Errorf("expected 4 checks, got %d", len(m.checks))
	}

	if m.checksLoaded != true {
		t.Error("expected checksLoaded=true")
	}
}

func TestUpdate_ReviewsLoaded(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	reviews := testReviews()
	msg := ReviewsLoadedMsg{Reviews: reviews}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if len(m.reviews) != 2 {
		t.Errorf("expected 2 reviews, got %d", len(m.reviews))
	}

	if m.reviewsLoaded != true {
		t.Error("expected reviewsLoaded=true")
	}
}

func TestUpdate_DiffLoaded(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	diff := "diff --git a/file.go b/file.go\n..."
	msg := DiffLoadedMsg{Diff: diff}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.diff != diff {
		t.Error("expected diff to be set")
	}

	if m.diffLoaded != true {
		t.Error("expected diffLoaded=true")
	}
}

func TestUpdate_LoadError(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	testErr := errors.New("load failed")
	msg := LoadErrorMsg{Err: testErr}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateError {
		t.Errorf("expected state=StateError, got %s", m.state)
	}

	if m.err != testErr {
		t.Errorf("expected err=%v, got %v", testErr, m.err)
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.width != 100 {
		t.Errorf("expected width=100, got %d", m.width)
	}

	if m.height != 50 {
		t.Errorf("expected height=50, got %d", m.height)
	}
}

func TestUpdate_KeyBack(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR()

	msg := tea.KeyMsg{Type: tea.KeyEsc}

	_, cmd := m.Update(msg)

	// Execute command to get the message
	if cmd == nil {
		t.Fatal("expected a command")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(BackToDashboardMsg); !ok {
		t.Errorf("expected BackToDashboardMsg, got %T", resultMsg)
	}
}

func TestUpdate_KeyTab(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR()
	m.prLoaded = true
	m.checksLoaded = true
	m.reviewsLoaded = true
	m.diffLoaded = true
	m.activeSection = SectionInfo

	msg := tea.KeyMsg{Type: tea.KeyTab}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.activeSection != SectionDescription {
		t.Errorf("expected SectionDescription, got %s", m.activeSection)
	}
}

func TestUpdate_KeyShiftTab(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR()
	m.prLoaded = true
	m.checksLoaded = true
	m.reviewsLoaded = true
	m.diffLoaded = true
	m.activeSection = SectionInfo

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.activeSection != SectionReviews {
		t.Errorf("expected SectionReviews, got %s", m.activeSection)
	}
}

func TestUpdate_KeyViewDiff(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}

	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected a command")
	}

	resultMsg := cmd()
	viewDiffMsg, ok := resultMsg.(ViewDiffMsg)
	if !ok {
		t.Fatalf("expected ViewDiffMsg, got %T", resultMsg)
	}

	if viewDiffMsg.PRNumber != 123 {
		t.Errorf("expected PRNumber=123, got %d", viewDiffMsg.PRNumber)
	}
}

func TestUpdate_KeyMerge(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR() // Mergeable PR

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateConfirming {
		t.Errorf("expected state=StateConfirming, got %s", m.state)
	}

	if m.confirmAction != ConfirmMerge {
		t.Errorf("expected confirmAction=ConfirmMerge, got %d", m.confirmAction)
	}
}

func TestUpdate_KeyClose(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR() // Open PR

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateConfirming {
		t.Errorf("expected state=StateConfirming, got %s", m.state)
	}

	if m.confirmAction != ConfirmClose {
		t.Errorf("expected confirmAction=ConfirmClose, got %d", m.confirmAction)
	}
}

func TestUpdate_SectionNumbers(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR()
	m.prLoaded = true
	m.checksLoaded = true
	m.reviewsLoaded = true
	m.diffLoaded = true

	tests := []struct {
		key      string
		expected Section
	}{
		{"1", SectionInfo},
		{"2", SectionDescription},
		{"3", SectionChecks},
		{"4", SectionReviews},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}

			newModel, _ := m.Update(msg)
			m = newModel.(Model)

			if m.activeSection != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, m.activeSection)
			}
		})
	}
}

// =============================================================================
// Section Rendering Tests
// =============================================================================

func TestRenderInfoSection(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.pr = testPR()
	m.checks = testChecks()
	m.reviews = testReviews()

	info := m.renderInfoSection()

	// Verify key content is present
	if !contains(info, "#123") {
		t.Error("expected PR number in info section")
	}

	if !contains(info, "Add new feature") {
		t.Error("expected PR title in info section")
	}

	if !contains(info, "testuser") {
		t.Error("expected author in info section")
	}

	if !contains(info, "feature/new-feature") {
		t.Error("expected head branch in info section")
	}

	if !contains(info, "main") {
		t.Error("expected base branch in info section")
	}

	if !contains(info, "+150") {
		t.Error("expected additions in info section")
	}

	if !contains(info, "-30") {
		t.Error("expected deletions in info section")
	}
}

func TestRenderDescriptionSection(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.pr = testPR()

	desc := m.renderDescriptionSection()

	if !contains(desc, "Description") {
		t.Error("expected Description header")
	}

	if !contains(desc, "This PR adds a new feature") {
		t.Error("expected PR body content")
	}
}

func TestRenderDescriptionSection_WithSummary(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.pr = testPR()
	m.summary = "This is an AI-generated summary of the PR."

	desc := m.renderDescriptionSection()

	if !contains(desc, "AI Summary") {
		t.Error("expected AI Summary header")
	}

	if !contains(desc, "AI-generated summary") {
		t.Error("expected summary content")
	}
}

func TestRenderDescriptionSection_Empty(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.pr = &gh.PR{
		Number: 123,
		Body:   "",
	}

	desc := m.renderDescriptionSection()

	if !contains(desc, "No description provided") {
		t.Error("expected empty description message")
	}
}

func TestRenderChecksSection(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.checks = testChecks()

	checks := m.renderChecksSection()

	if !contains(checks, "CI Checks") {
		t.Error("expected CI Checks header")
	}

	if !contains(checks, "build") {
		t.Error("expected build check")
	}

	if !contains(checks, "lint") {
		t.Error("expected lint check")
	}

	if !contains(checks, "test") {
		t.Error("expected test check")
	}
}

func TestRenderChecksSection_Empty(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.checks = []gh.Check{}

	checks := m.renderChecksSection()

	if !contains(checks, "No checks found") {
		t.Error("expected empty checks message")
	}
}

func TestRenderReviewsSection(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.pr = testPR()
	m.reviews = testReviews()

	reviews := m.renderReviewsSection()

	if !contains(reviews, "Reviews") {
		t.Error("expected Reviews header")
	}

	if !contains(reviews, "reviewer1") {
		t.Error("expected reviewer1")
	}

	if !contains(reviews, "reviewer2") {
		t.Error("expected reviewer2")
	}
}

func TestRenderReviewsSection_Empty(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.pr = &gh.PR{Reviewers: []gh.Reviewer{}}
	m.reviews = []gh.Review{}

	reviews := m.renderReviewsSection()

	if !contains(reviews, "No reviews submitted") {
		t.Error("expected empty reviews message")
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestView_Loading(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateLoading
	m.width = 80
	m.height = 24

	view := m.View()

	// Loading view should contain spinner
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestView_Error(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateError
	m.err = errors.New("test error")
	m.width = 80
	m.height = 24

	view := m.View()

	if !contains(view, "Error") {
		t.Error("expected Error in view")
	}

	if !contains(view, "retry") || !contains(view, "back") {
		t.Error("expected help hints in error view")
	}
}

func TestView_Ready(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR()
	m.checks = testChecks()
	m.reviews = testReviews()
	m.width = 80
	m.height = 24

	view := m.View()

	// Should contain header elements
	if !contains(view, "#123") {
		t.Error("expected PR number in view")
	}

	// Should contain tab bar
	if !contains(view, "Info") {
		t.Error("expected Info tab in view")
	}

	// Should contain footer/help
	if !contains(view, "esc") {
		t.Error("expected esc key hint in view")
	}
}

// =============================================================================
// State String Tests
// =============================================================================

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateLoading, "loading"},
		{StateReady, "ready"},
		{StateError, "error"},
		{StateSummarizing, "summarizing"},
		{StateMerging, "merging"},
		{StateClosing, "closing"},
		{StateConfirming, "confirming"},
		{State(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.state.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.state.String())
			}
		})
	}
}

func TestSectionString(t *testing.T) {
	tests := []struct {
		section  Section
		expected string
	}{
		{SectionInfo, "Info"},
		{SectionDescription, "Description"},
		{SectionChecks, "Checks"},
		{SectionReviews, "Reviews"},
		{Section(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.section.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.section.String())
			}
		})
	}
}

// =============================================================================
// Check Status Mapping Tests
// =============================================================================

func TestMapCheckStatus(t *testing.T) {
	tests := []struct {
		name     string
		check    gh.Check
		expected ui.CheckStatus
	}{
		{
			name:     "success conclusion",
			check:    gh.Check{Status: "completed", Conclusion: "success"},
			expected: ui.CheckSuccess,
		},
		{
			name:     "failure conclusion",
			check:    gh.Check{Status: "completed", Conclusion: "failure"},
			expected: ui.CheckFailure,
		},
		{
			name:     "skipped conclusion",
			check:    gh.Check{Status: "completed", Conclusion: "skipped"},
			expected: ui.CheckSkipped,
		},
		{
			name:     "cancelled conclusion",
			check:    gh.Check{Status: "completed", Conclusion: "cancelled"},
			expected: ui.CheckCancelled,
		},
		{
			name:     "in_progress status",
			check:    gh.Check{Status: "in_progress", Conclusion: ""},
			expected: ui.CheckRunning,
		},
		{
			name:     "queued status",
			check:    gh.Check{Status: "queued", Conclusion: ""},
			expected: ui.CheckPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := mapCheckStatus(tt.check)
			if status != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, status)
			}
		})
	}
}

func TestCountCheckStatus(t *testing.T) {
	checks := testChecks()

	passed, failed, pending := countCheckStatus(checks)

	if passed != 2 {
		t.Errorf("expected passed=2, got %d", passed)
	}

	if failed != 1 {
		t.Errorf("expected failed=1, got %d", failed)
	}

	if pending != 1 {
		t.Errorf("expected pending=1, got %d", pending)
	}
}

// =============================================================================
// LLM Integration Tests
// =============================================================================

func TestUpdate_LLMSummary(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateReady
	m.pr = testPR()
	m.diff = "diff content"

	// Set up mock LLM provider
	mockProvider := newMockLLMProvider([]string{"This ", "is ", "a ", "summary."}, nil)
	m.SetLLMProvider(mockProvider)

	// Trigger summary generation via key press
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if cmd == nil {
		t.Fatal("expected a command")
	}

	// Execute to get SummaryStartMsg
	resultMsg := cmd()
	if _, ok := resultMsg.(SummaryStartMsg); !ok {
		t.Fatalf("expected SummaryStartMsg, got %T", resultMsg)
	}
}

func TestUpdate_SummaryTokenMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateSummarizing
	m.pr = testPR()

	// Set up channels for streaming
	tokenCh := make(chan llm.Token, 10)
	errCh := make(chan error, 1)
	m.tokenCh = tokenCh
	m.errCh = errCh

	// Send a token
	msg := SummaryTokenMsg{Token: "Hello "}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.summary != "Hello " {
		t.Errorf("expected summary='Hello ', got %q", m.summary)
	}

	// Send another token
	msg = SummaryTokenMsg{Token: "World"}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.summary != "Hello World" {
		t.Errorf("expected summary='Hello World', got %q", m.summary)
	}
}

func TestUpdate_SummaryDoneMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateSummarizing
	m.summary = "Complete summary"

	msg := SummaryDoneMsg{}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateReady {
		t.Errorf("expected state=StateReady, got %s", m.state)
	}

	if m.summary != "Complete summary" {
		t.Error("summary should be preserved")
	}
}

func TestUpdate_SummaryErrorMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateSummarizing

	testErr := errors.New("LLM failed")
	msg := SummaryErrorMsg{Err: testErr}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateError {
		t.Errorf("expected state=StateError, got %s", m.state)
	}

	if m.err != testErr {
		t.Errorf("expected err=%v, got %v", testErr, m.err)
	}
}

// =============================================================================
// Action Result Tests
// =============================================================================

func TestUpdate_MergeDoneMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateMerging

	msg := MergeDoneMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateReady {
		t.Errorf("expected state=StateReady, got %s", m.state)
	}

	// Should reload PR
	if cmd == nil {
		t.Error("expected reload command")
	}
}

func TestUpdate_MergeErrorMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateMerging

	testErr := errors.New("merge failed")
	msg := MergeErrorMsg{Err: testErr}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateError {
		t.Errorf("expected state=StateError, got %s", m.state)
	}

	if m.err != testErr {
		t.Errorf("expected err=%v, got %v", testErr, m.err)
	}
}

func TestUpdate_CloseDoneMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateClosing

	msg := CloseDoneMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateReady {
		t.Errorf("expected state=StateReady, got %s", m.state)
	}

	// Should reload PR
	if cmd == nil {
		t.Error("expected reload command")
	}
}

func TestUpdate_CloseErrorMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.state = StateClosing

	testErr := errors.New("close failed")
	msg := CloseErrorMsg{Err: testErr}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.state != StateError {
		t.Errorf("expected state=StateError, got %s", m.state)
	}

	if m.err != testErr {
		t.Errorf("expected err=%v, got %v", testErr, m.err)
	}
}

// =============================================================================
// Cleanup Tests
// =============================================================================

func TestCleanup(t *testing.T) {
	t.Run("cancels context and clears channels", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.pr = testPR()
		m.diff = "diff content"

		// Simulate LLM streaming state
		ctx, cancel := context.WithCancel(context.Background())
		m.llmCtx = ctx
		m.llmCancel = cancel
		tokenCh := make(chan llm.Token)
		errCh := make(chan error)
		m.tokenCh = tokenCh
		m.errCh = errCh

		// Call Cleanup
		m.Cleanup()

		// Verify everything is cleaned up
		if m.llmCancel != nil {
			t.Error("expected llmCancel to be nil after cleanup")
		}
		if m.llmCtx != nil {
			t.Error("expected llmCtx to be nil after cleanup")
		}
		if m.tokenCh != nil {
			t.Error("expected tokenCh to be nil after cleanup")
		}
		if m.errCh != nil {
			t.Error("expected errCh to be nil after cleanup")
		}

		// Verify context was cancelled
		select {
		case <-ctx.Done():
			// Expected - context was cancelled
		default:
			t.Error("expected context to be cancelled")
		}
	})

	t.Run("is safe to call multiple times", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)

		// Call Cleanup multiple times - should not panic
		m.Cleanup()
		m.Cleanup()
		m.Cleanup()
	})

	t.Run("is safe to call with nil fields", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		// All fields are nil by default

		// Should not panic
		m.Cleanup()
	})
}

func TestCleanup_OnNavigation(t *testing.T) {
	t.Run("cleanup is called on back navigation", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.state = StateReady
		m.pr = testPR()

		// Set up LLM streaming state
		ctx, cancel := context.WithCancel(context.Background())
		m.llmCtx = ctx
		m.llmCancel = cancel

		// Trigger back navigation
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		newModel, cmd := m.Update(msg)
		m = newModel.(Model)

		// Verify cleanup was called
		if m.llmCancel != nil {
			t.Error("expected llmCancel to be nil after back navigation")
		}

		// Verify we're navigating back
		if cmd == nil {
			t.Fatal("expected a command")
		}
		resultMsg := cmd()
		if _, ok := resultMsg.(BackToDashboardMsg); !ok {
			t.Errorf("expected BackToDashboardMsg, got %T", resultMsg)
		}

		// Verify original context was cancelled
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Error("expected context to be cancelled on navigation")
		}
	})

	t.Run("cleanup is called on view diff navigation", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.state = StateReady
		m.pr = testPR()

		// Set up LLM streaming state
		ctx, cancel := context.WithCancel(context.Background())
		m.llmCtx = ctx
		m.llmCancel = cancel

		// Trigger view diff navigation
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		newModel, cmd := m.Update(msg)
		m = newModel.(Model)

		// Verify cleanup was called
		if m.llmCancel != nil {
			t.Error("expected llmCancel to be nil after view diff navigation")
		}

		// Verify we're navigating to diff view
		if cmd == nil {
			t.Fatal("expected a command")
		}
		resultMsg := cmd()
		if _, ok := resultMsg.(ViewDiffMsg); !ok {
			t.Errorf("expected ViewDiffMsg, got %T", resultMsg)
		}

		// Verify original context was cancelled
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Error("expected context to be cancelled on navigation")
		}
	})

	t.Run("cleanup is called during summarizing state on cancel", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.state = StateSummarizing
		m.pr = testPR()

		// Set up LLM streaming state
		ctx, cancel := context.WithCancel(context.Background())
		m.llmCtx = ctx
		m.llmCancel = cancel

		// Trigger cancel during summarization (ESC key cancels but stays on screen)
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		// Verify cleanup was called
		if m.llmCancel != nil {
			t.Error("expected llmCancel to be nil after cancel during summarization")
		}

		// Verify context was cancelled
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Error("expected context to be cancelled on cancel")
		}

		// Verify state is back to ready (cancellation stops summarizing but stays on screen)
		if m.state != StateReady {
			t.Errorf("expected state=StateReady after cancel, got %s", m.state)
		}
	})

	t.Run("cleanup is called on start review navigation", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.state = StateReady
		m.pr = testPR()
		m.diff = "diff content"

		// Set up LLM streaming state
		ctx, cancel := context.WithCancel(context.Background())
		m.llmCtx = ctx
		m.llmCancel = cancel

		// Trigger start review navigation
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}}
		newModel, cmd := m.Update(msg)
		m = newModel.(Model)

		// Verify cleanup was called
		if m.llmCancel != nil {
			t.Error("expected llmCancel to be nil after start review navigation")
		}

		// Verify we're navigating to review
		if cmd == nil {
			t.Fatal("expected a command")
		}
		resultMsg := cmd()
		if _, ok := resultMsg.(StartReviewMsg); !ok {
			t.Errorf("expected StartReviewMsg, got %T", resultMsg)
		}

		// Verify original context was cancelled
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Error("expected context to be cancelled on navigation")
		}
	})
}

// =============================================================================
// Stale Message Detection Tests
// =============================================================================

func TestGenerationID(t *testing.T) {
	t.Run("initial generation ID is zero", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)

		if m.GenerationID() != 0 {
			t.Errorf("expected initial generationID=0, got %d", m.GenerationID())
		}
	})

	t.Run("generation ID increments on summary start", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.state = StateReady
		m.pr = testPR()
		m.diff = "diff content"

		// Set up mock LLM provider
		mockProvider := newMockLLMProvider([]string{"test"}, nil)
		m.SetLLMProvider(mockProvider)

		initialID := m.GenerationID()

		// Trigger summary start
		msg := SummaryStartMsg{}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if m.GenerationID() != initialID+1 {
			t.Errorf("expected generationID=%d, got %d", initialID+1, m.GenerationID())
		}
	})
}

func TestStaleMessageDetection(t *testing.T) {
	t.Run("stale tokens are ignored", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.state = StateSummarizing
		m.pr = testPR()
		m.generationID = 5 // Current generation

		// Set up channels for streaming (to prevent readNextToken from panicking)
		tokenCh := make(chan llm.Token, 10)
		errCh := make(chan error, 1)
		m.tokenCh = tokenCh
		m.errCh = errCh

		// Send a stale token (from generation 3)
		staleMsg := SummaryTokenMsg{Token: "stale token", GenerationID: 3}
		newModel, cmd := m.Update(staleMsg)
		m = newModel.(Model)

		// Verify token was ignored
		if m.summary != "" {
			t.Errorf("expected summary to be empty (stale token ignored), got %q", m.summary)
		}

		// Verify no command was returned (no further token reading)
		if cmd != nil {
			t.Error("expected no command for stale token")
		}
	})

	t.Run("current generation tokens are processed", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.state = StateSummarizing
		m.pr = testPR()
		m.generationID = 5 // Current generation

		// Set up channels for streaming
		tokenCh := make(chan llm.Token, 10)
		errCh := make(chan error, 1)
		m.tokenCh = tokenCh
		m.errCh = errCh

		// Send a token from the current generation
		currentMsg := SummaryTokenMsg{Token: "valid token", GenerationID: 5}
		newModel, cmd := m.Update(currentMsg)
		m = newModel.(Model)

		// Verify token was processed
		if m.summary != "valid token" {
			t.Errorf("expected summary='valid token', got %q", m.summary)
		}

		// Verify a command was returned (to read next token)
		if cmd == nil {
			t.Error("expected a command to read next token")
		}
	})
}

func TestDoubleSummarizationPrevention(t *testing.T) {
	t.Run("cannot start summarization while already summarizing", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, 123)
		m.state = StateSummarizing
		m.pr = testPR()
		m.diff = "diff content"
		m.summary = "existing partial summary"

		// Set up mock LLM provider
		mockProvider := newMockLLMProvider([]string{"new token"}, nil)
		m.SetLLMProvider(mockProvider)

		initialGenID := m.generationID

		// Try to start another summarization
		msg := SummaryStartMsg{}
		newModel, cmd := m.Update(msg)
		m = newModel.(Model)

		// Verify state didn't change
		if m.state != StateSummarizing {
			t.Errorf("expected state to remain StateSummarizing, got %s", m.state)
		}

		// Verify generation ID didn't increment
		if m.generationID != initialGenID {
			t.Errorf("expected generationID to remain %d, got %d", initialGenID, m.generationID)
		}

		// Verify summary wasn't cleared
		if m.summary != "existing partial summary" {
			t.Errorf("expected summary to remain 'existing partial summary', got %q", m.summary)
		}

		// Verify no command was returned
		if cmd != nil {
			t.Error("expected no command when already summarizing")
		}
	})
}

// =============================================================================
// Helpers
// =============================================================================

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
