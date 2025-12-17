package root

import (
	"testing"

	"bui/internal/app/createpr"
	"bui/internal/app/dashboard"
	"bui/internal/app/diffview"
	"bui/internal/app/prdetail"
	"bui/internal/app/review"
	"bui/internal/config"
	"bui/internal/gh"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Test Helpers
// =============================================================================

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
			AutoRefresh: false,
		},
	}
}

func testPR() gh.PR {
	return gh.PR{
		Number:     42,
		Title:      "Test PR",
		Body:       "Test body",
		State:      "open",
		Author:     gh.User{Login: "testuser"},
		HeadBranch: "feature/test",
		BaseBranch: "main",
		URL:        "https://github.com/test/repo/pull/42",
		Mergeable:  "MERGEABLE",
	}
}

// =============================================================================
// Screen Type Tests
// =============================================================================

func TestScreen_String(t *testing.T) {
	tests := []struct {
		name   string
		screen Screen
		want   string
	}{
		{name: "dashboard", screen: ScreenDashboard, want: "dashboard"},
		{name: "prdetail", screen: ScreenPRDetail, want: "prdetail"},
		{name: "diffview", screen: ScreenDiffView, want: "diffview"},
		{name: "createpr", screen: ScreenCreatePR, want: "createpr"},
		{name: "review", screen: ScreenReview, want: "review"},
		{name: "unknown", screen: Screen(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.screen.String(); got != tt.want {
				t.Errorf("Screen.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Model Initialization Tests
// =============================================================================

func TestNew(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "/path/to/config.toml")

	if model.screen != ScreenDashboard {
		t.Errorf("New() screen = %v, want %v", model.screen, ScreenDashboard)
	}

	if model.cfgPath != "/path/to/config.toml" {
		t.Errorf("New() cfgPath = %q, want %q", model.cfgPath, "/path/to/config.toml")
	}

	if model.width != 0 || model.height != 0 {
		t.Errorf("New() dimensions = (%d, %d), want (0, 0)", model.width, model.height)
	}
}

func TestModel_Init(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	cmd := model.Init()

	// Init should return a command to initialize the dashboard
	if cmd == nil {
		t.Error("Init() returned nil, want command to initialize dashboard")
	}
}

// =============================================================================
// Accessor Tests
// =============================================================================

func TestModel_CurrentScreen(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	if got := model.CurrentScreen(); got != ScreenDashboard {
		t.Errorf("CurrentScreen() = %v, want %v", got, ScreenDashboard)
	}
}

func TestModel_SetScreen(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	model.SetScreen(ScreenPRDetail)

	if model.screen != ScreenPRDetail {
		t.Errorf("SetScreen(ScreenPRDetail) screen = %v, want %v", model.screen, ScreenPRDetail)
	}
}

func TestModel_Dimensions(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Set dimensions via WindowSizeMsg
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m := newModel.(Model)

	if m.Width() != 120 {
		t.Errorf("Width() = %d, want %d", m.Width(), 120)
	}

	if m.Height() != 40 {
		t.Errorf("Height() = %d, want %d", m.Height(), 40)
	}
}

func TestModel_Dashboard(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Dashboard should be initialized
	dash := model.Dashboard()

	// Verify it's a valid dashboard model (non-nil)
	_ = dash.View()
}

func TestModel_PRDetail(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to PR detail first
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	prDetail := m.PRDetail()

	if prDetail.PRNumber() != 42 {
		t.Errorf("PRDetail().PRNumber() = %d, want %d", prDetail.PRNumber(), 42)
	}
}

func TestModel_DiffView(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to diff view first
	newModel, _ := model.Update(prdetail.ViewDiffMsg{PRNumber: 42})
	m := newModel.(Model)

	diffView := m.DiffView()

	if diffView.PRNumber() != 42 {
		t.Errorf("DiffView().PRNumber() = %d, want %d", diffView.PRNumber(), 42)
	}
}

// =============================================================================
// WindowSizeMsg Tests
// =============================================================================

func TestModel_Update_WindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.width != 100 {
		t.Errorf("Update(WindowSizeMsg) width = %d, want %d", m.width, 100)
	}

	if m.height != 50 {
		t.Errorf("Update(WindowSizeMsg) height = %d, want %d", m.height, 50)
	}
}

func TestModel_Update_WindowSizeMsg_PropagatesSize(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Set initial size
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := newModel.(Model)

	// Navigate to PR detail
	newModel, _ = m.Update(dashboard.OpenPRDetailMsg{PR: testPR()})
	m = newModel.(Model)

	// Update size while on PR detail screen
	newModel, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = newModel.(Model)

	// Verify dimensions were updated
	if m.width != 120 || m.height != 40 {
		t.Errorf("Update(WindowSizeMsg) on PRDetail dimensions = (%d, %d), want (120, 40)",
			m.width, m.height)
	}
}

// =============================================================================
// Global Quit Handler Tests
// =============================================================================

func TestModel_Update_Quit_QKey(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Update(KeyMsg 'q') returned nil cmd, want tea.Quit")
	}
}

func TestModel_Update_Quit_CtrlC(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Update(KeyMsg ctrl+c) returned nil cmd, want tea.Quit")
	}
}

func TestModel_Update_Quit_FromAnyScreen(t *testing.T) {
	screens := []Screen{ScreenDashboard, ScreenPRDetail, ScreenDiffView, ScreenCreatePR, ScreenReview}

	for _, screen := range screens {
		t.Run(screen.String(), func(t *testing.T) {
			cfg := testConfig()
			model := New(cfg, "")
			model.SetScreen(screen)

			// Need to initialize the screen models for non-dashboard screens
			switch screen {
			case ScreenPRDetail:
				model.prdetail = prdetail.New(cfg, 1)
			case ScreenDiffView:
				model.diffview = diffview.New(cfg, 1)
			case ScreenCreatePR:
				model.createpr = createpr.New(cfg, nil, nil, nil)
			case ScreenReview:
				model.review = review.New(cfg, nil, nil, 1, "Test PR", "diff")
			}

			msg := tea.KeyMsg{Type: tea.KeyCtrlC}

			_, cmd := model.Update(msg)

			if cmd == nil {
				t.Errorf("Update(ctrl+c) from %s returned nil cmd, want tea.Quit", screen)
			}
		})
	}
}

// =============================================================================
// Navigation Tests: Dashboard -> PR Detail
// =============================================================================

func TestModel_Update_OpenPRDetailMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	pr := testPR()
	msg := dashboard.OpenPRDetailMsg{PR: pr}

	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.screen != ScreenPRDetail {
		t.Errorf("Update(OpenPRDetailMsg) screen = %v, want %v", m.screen, ScreenPRDetail)
	}

	if m.prdetail.PRNumber() != pr.Number {
		t.Errorf("Update(OpenPRDetailMsg) prdetail.PRNumber() = %d, want %d",
			m.prdetail.PRNumber(), pr.Number)
	}

	// Should return a command to initialize the PR detail screen
	if cmd == nil {
		t.Error("Update(OpenPRDetailMsg) returned nil cmd, want Init command")
	}
}

// =============================================================================
// Navigation Tests: PR Detail -> Dashboard
// =============================================================================

func TestModel_Update_BackToDashboardMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to PR detail first
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	if m.screen != ScreenPRDetail {
		t.Fatalf("Setup failed: screen = %v, want %v", m.screen, ScreenPRDetail)
	}

	// Navigate back to dashboard
	msg := prdetail.BackToDashboardMsg{}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.screen != ScreenDashboard {
		t.Errorf("Update(BackToDashboardMsg) screen = %v, want %v", m.screen, ScreenDashboard)
	}

	// Should not return a command (dashboard state is preserved)
	if cmd != nil {
		t.Error("Update(BackToDashboardMsg) returned non-nil cmd, want nil")
	}
}

// =============================================================================
// Navigation Tests: PR Detail -> Diff View
// =============================================================================

func TestModel_Update_ViewDiffMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate to PR detail first
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	// Navigate to diff view
	msg := prdetail.ViewDiffMsg{PRNumber: pr.Number}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.screen != ScreenDiffView {
		t.Errorf("Update(ViewDiffMsg) screen = %v, want %v", m.screen, ScreenDiffView)
	}

	if m.diffview.PRNumber() != pr.Number {
		t.Errorf("Update(ViewDiffMsg) diffview.PRNumber() = %d, want %d",
			m.diffview.PRNumber(), pr.Number)
	}

	// Should return a command to initialize the diff view screen
	if cmd == nil {
		t.Error("Update(ViewDiffMsg) returned nil cmd, want Init command")
	}
}

// =============================================================================
// Navigation Tests: Diff View -> PR Detail
// =============================================================================

func TestModel_Update_BackToPRDetailMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to PR detail first
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	// Navigate to diff view
	newModel, _ = m.Update(prdetail.ViewDiffMsg{PRNumber: pr.Number})
	m = newModel.(Model)

	if m.screen != ScreenDiffView {
		t.Fatalf("Setup failed: screen = %v, want %v", m.screen, ScreenDiffView)
	}

	// Navigate back to PR detail
	msg := diffview.BackToPRDetailMsg{PRNumber: pr.Number}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.screen != ScreenPRDetail {
		t.Errorf("Update(BackToPRDetailMsg) screen = %v, want %v", m.screen, ScreenPRDetail)
	}

	// Should not return a command (PR detail state is preserved)
	if cmd != nil {
		t.Error("Update(BackToPRDetailMsg) returned non-nil cmd, want nil")
	}
}

// =============================================================================
// Navigation Tests: Dashboard -> Create PR
// =============================================================================

func TestModel_Update_CreatePRMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	msg := dashboard.CreatePRMsg{}

	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	// Should navigate to create PR screen
	if m.screen != ScreenCreatePR {
		t.Errorf("Update(CreatePRMsg) screen = %v, want %v", m.screen, ScreenCreatePR)
	}

	// Should return a command to initialize the create PR screen
	if cmd == nil {
		t.Error("Update(CreatePRMsg) returned nil cmd, want Init command")
	}
}

// =============================================================================
// Navigation Tests: Create PR -> Dashboard
// =============================================================================

func TestModel_Update_CreatePR_BackToDashboardMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to create PR first
	newModel, _ := model.Update(dashboard.CreatePRMsg{})
	m := newModel.(Model)

	if m.screen != ScreenCreatePR {
		t.Fatalf("Setup failed: screen = %v, want %v", m.screen, ScreenCreatePR)
	}

	// Navigate back to dashboard
	msg := createpr.BackToDashboardMsg{}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.screen != ScreenDashboard {
		t.Errorf("Update(createpr.BackToDashboardMsg) screen = %v, want %v", m.screen, ScreenDashboard)
	}

	// Should not return a command (dashboard state is preserved)
	if cmd != nil {
		t.Error("Update(createpr.BackToDashboardMsg) returned non-nil cmd, want nil")
	}
}

// =============================================================================
// Navigation Tests: Create PR -> PR Detail (after creation)
// =============================================================================

func TestModel_Update_PRCreatedMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate to create PR first
	newModel, _ := model.Update(dashboard.CreatePRMsg{})
	m := newModel.(Model)

	// Simulate PR creation
	newPR := &gh.PR{Number: 123, Title: "New PR"}
	msg := createpr.PRCreatedMsg{PR: newPR}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.screen != ScreenPRDetail {
		t.Errorf("Update(PRCreatedMsg) screen = %v, want %v", m.screen, ScreenPRDetail)
	}

	if m.prdetail.PRNumber() != 123 {
		t.Errorf("Update(PRCreatedMsg) prdetail.PRNumber() = %d, want %d", m.prdetail.PRNumber(), 123)
	}

	// Should return a command to initialize the PR detail screen
	if cmd == nil {
		t.Error("Update(PRCreatedMsg) returned nil cmd, want Init command")
	}
}

// =============================================================================
// Navigation Tests: PR Detail -> Review
// =============================================================================

func TestModel_Update_StartReviewMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate to PR detail first
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	// Navigate to review
	msg := prdetail.StartReviewMsg{
		PRNumber: pr.Number,
		PRTitle:  pr.Title,
		Diff:     "diff content",
	}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.screen != ScreenReview {
		t.Errorf("Update(StartReviewMsg) screen = %v, want %v", m.screen, ScreenReview)
	}

	if m.review.PRNumber() != pr.Number {
		t.Errorf("Update(StartReviewMsg) review.PRNumber() = %d, want %d", m.review.PRNumber(), pr.Number)
	}

	// Should return a command to initialize the review screen
	if cmd == nil {
		t.Error("Update(StartReviewMsg) returned nil cmd, want Init command")
	}
}

// =============================================================================
// Navigation Tests: Review -> PR Detail
// =============================================================================

func TestModel_Update_Review_BackToPRDetailMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate to PR detail first
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	// Navigate to review
	newModel, _ = m.Update(prdetail.StartReviewMsg{PRNumber: pr.Number, PRTitle: pr.Title, Diff: "diff"})
	m = newModel.(Model)

	if m.screen != ScreenReview {
		t.Fatalf("Setup failed: screen = %v, want %v", m.screen, ScreenReview)
	}

	// Navigate back to PR detail
	msg := review.BackToPRDetailMsg{PRNumber: pr.Number}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.screen != ScreenPRDetail {
		t.Errorf("Update(review.BackToPRDetailMsg) screen = %v, want %v", m.screen, ScreenPRDetail)
	}

	// Should not return a command (PR detail state is preserved)
	if cmd != nil {
		t.Error("Update(review.BackToPRDetailMsg) returned non-nil cmd, want nil")
	}
}

func TestModel_Update_ReviewSubmittedMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate to PR detail first
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	// Navigate to review
	newModel, _ = m.Update(prdetail.StartReviewMsg{PRNumber: pr.Number, PRTitle: pr.Title, Diff: "diff"})
	m = newModel.(Model)

	// Submit review
	msg := review.ReviewSubmittedMsg{}

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	if m.screen != ScreenPRDetail {
		t.Errorf("Update(ReviewSubmittedMsg) screen = %v, want %v", m.screen, ScreenPRDetail)
	}

	// Should return a command to refresh PR detail
	if cmd == nil {
		t.Error("Update(ReviewSubmittedMsg) returned nil cmd, want Init command")
	}
}

// =============================================================================
// Message Routing Tests
// =============================================================================

func TestModel_Update_RoutesToDashboard(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.SetScreen(ScreenDashboard)

	// Send a key that dashboard should handle
	msg := tea.KeyMsg{Type: tea.KeyDown}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Verify we're still on dashboard and it processed the message
	if m.screen != ScreenDashboard {
		t.Errorf("Message routing changed screen unexpectedly: got %v", m.screen)
	}
}

func TestModel_Update_RoutesToPRDetail(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to PR detail
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	// Send a key that PR detail should handle
	msg := tea.KeyMsg{Type: tea.KeyDown}

	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Verify we're still on PR detail
	if m.screen != ScreenPRDetail {
		t.Errorf("Message routing changed screen unexpectedly: got %v", m.screen)
	}
}

func TestModel_Update_RoutesToDiffView(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to diff view
	newModel, _ := model.Update(prdetail.ViewDiffMsg{PRNumber: 42})
	m := newModel.(Model)

	// Send a key that diff view should handle
	msg := tea.KeyMsg{Type: tea.KeyDown}

	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Verify we're still on diff view
	if m.screen != ScreenDiffView {
		t.Errorf("Message routing changed screen unexpectedly: got %v", m.screen)
	}
}

func TestModel_Update_RoutesToCreatePR(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to create PR
	newModel, _ := model.Update(dashboard.CreatePRMsg{})
	m := newModel.(Model)

	// Send a key that create PR should handle
	msg := tea.KeyMsg{Type: tea.KeyDown}

	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Verify we're still on create PR
	if m.screen != ScreenCreatePR {
		t.Errorf("Message routing changed screen unexpectedly: got %v", m.screen)
	}
}

func TestModel_Update_RoutesToReview(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to review
	newModel, _ := model.Update(prdetail.StartReviewMsg{PRNumber: 42, PRTitle: "Test", Diff: "diff"})
	m := newModel.(Model)

	// Send a key that review should handle
	msg := tea.KeyMsg{Type: tea.KeyDown}

	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	// Verify we're still on review
	if m.screen != ScreenReview {
		t.Errorf("Message routing changed screen unexpectedly: got %v", m.screen)
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestModel_View_Dashboard(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.SetScreen(ScreenDashboard)

	// Set size for rendering
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := newModel.(Model)

	view := m.View()

	if view == "" {
		t.Error("View() on Dashboard returned empty string")
	}
}

func TestModel_View_PRDetail(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to PR detail
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	// Set size
	newModel, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	view := m.View()

	if view == "" {
		t.Error("View() on PRDetail returned empty string")
	}
}

func TestModel_View_DiffView(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to diff view
	newModel, _ := model.Update(prdetail.ViewDiffMsg{PRNumber: 42})
	m := newModel.(Model)

	// Set size
	newModel, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	view := m.View()

	if view == "" {
		t.Error("View() on DiffView returned empty string")
	}
}

func TestModel_View_CreatePR(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to create PR
	newModel, _ := model.Update(dashboard.CreatePRMsg{})
	m := newModel.(Model)

	// Set size
	newModel, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	view := m.View()

	if view == "" {
		t.Error("View() on CreatePR returned empty string")
	}
}

func TestModel_View_Review(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Navigate to review
	newModel, _ := model.Update(prdetail.StartReviewMsg{PRNumber: 42, PRTitle: "Test", Diff: "diff"})
	m := newModel.(Model)

	// Set size
	newModel, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	view := m.View()

	if view == "" {
		t.Error("View() on Review returned empty string")
	}
}

func TestModel_View_UnknownScreen(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.screen = Screen(99) // Invalid screen

	view := model.View()

	if view != "Unknown screen" {
		t.Errorf("View() on unknown screen = %q, want %q", view, "Unknown screen")
	}
}

// =============================================================================
// Full Navigation Workflow Test
// =============================================================================

func TestModel_FullNavigationWorkflow(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// 1. Start on dashboard
	if model.CurrentScreen() != ScreenDashboard {
		t.Fatalf("Initial screen = %v, want %v", model.CurrentScreen(), ScreenDashboard)
	}

	// 2. Set initial size
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m := newModel.(Model)

	// 3. Navigate to PR detail
	pr := testPR()
	newModel, cmd := m.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m = newModel.(Model)

	if m.CurrentScreen() != ScreenPRDetail {
		t.Errorf("After OpenPRDetailMsg, screen = %v, want %v", m.CurrentScreen(), ScreenPRDetail)
	}
	if cmd == nil {
		t.Error("OpenPRDetailMsg should return Init command")
	}

	// 4. Navigate to diff view
	newModel, cmd = m.Update(prdetail.ViewDiffMsg{PRNumber: pr.Number})
	m = newModel.(Model)

	if m.CurrentScreen() != ScreenDiffView {
		t.Errorf("After ViewDiffMsg, screen = %v, want %v", m.CurrentScreen(), ScreenDiffView)
	}
	if cmd == nil {
		t.Error("ViewDiffMsg should return Init command")
	}

	// 5. Navigate back to PR detail
	newModel, _ = m.Update(diffview.BackToPRDetailMsg{PRNumber: pr.Number})
	m = newModel.(Model)

	if m.CurrentScreen() != ScreenPRDetail {
		t.Errorf("After BackToPRDetailMsg, screen = %v, want %v", m.CurrentScreen(), ScreenPRDetail)
	}

	// 6. Navigate back to dashboard
	newModel, _ = m.Update(prdetail.BackToDashboardMsg{})
	m = newModel.(Model)

	if m.CurrentScreen() != ScreenDashboard {
		t.Errorf("After BackToDashboardMsg, screen = %v, want %v", m.CurrentScreen(), ScreenDashboard)
	}

	// 7. Verify views render without panic
	_ = m.View()
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestModel_Update_MultipleWindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Send multiple size updates
	sizes := []struct{ w, h int }{
		{80, 24},
		{120, 40},
		{200, 60},
		{80, 24},
	}

	m := model
	for _, size := range sizes {
		newModel, _ := m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
		m = newModel.(Model)

		if m.Width() != size.w || m.Height() != size.h {
			t.Errorf("After resize to (%d, %d), dimensions = (%d, %d)",
				size.w, size.h, m.Width(), m.Height())
		}
	}
}

func TestModel_Update_NavigationPreservesState(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate to PR detail
	pr := testPR()
	newModel, _ := model.Update(dashboard.OpenPRDetailMsg{PR: pr})
	m := newModel.(Model)

	// Navigate to diff view
	newModel, _ = m.Update(prdetail.ViewDiffMsg{PRNumber: pr.Number})
	m = newModel.(Model)

	// Navigate back to PR detail - should preserve state
	newModel, _ = m.Update(diffview.BackToPRDetailMsg{PRNumber: pr.Number})
	m = newModel.(Model)

	// Verify PR detail still has the PR number
	if m.prdetail.PRNumber() != pr.Number {
		t.Errorf("PR detail lost state: PRNumber = %d, want %d",
			m.prdetail.PRNumber(), pr.Number)
	}
}

func TestModel_Update_ZeroDimensions(t *testing.T) {
	cfg := testConfig()
	model := New(cfg, "")

	// Set zero dimensions
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 0, Height: 0})
	m := newModel.(Model)

	// Should handle gracefully
	if m.Width() != 0 || m.Height() != 0 {
		t.Errorf("Zero dimensions not preserved: got (%d, %d)", m.Width(), m.Height())
	}

	// View should not panic
	_ = m.View()
}
