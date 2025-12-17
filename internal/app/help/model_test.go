package help

import (
	"strings"
	"testing"

	"bui/internal/config"

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
	}
}

// =============================================================================
// HelpContext Tests
// =============================================================================

func TestHelpContextString(t *testing.T) {
	tests := []struct {
		context  HelpContext
		expected string
	}{
		{HelpContextGlobal, "Global"},
		{HelpContextDashboard, "Dashboard"},
		{HelpContextPRDetail, "PR Detail"},
		{HelpContextDiffView, "Diff View"},
		{HelpContextCreatePR, "Create PR"},
		{HelpContextReview, "Review"},
		{HelpContextComposer, "Composer"},
		{HelpContext(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.context.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.context.String())
			}
		})
	}
}

// =============================================================================
// Model Tests
// =============================================================================

func TestNew(t *testing.T) {
	t.Run("creates model with global context", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, HelpContextGlobal)

		if m.Context() != HelpContextGlobal {
			t.Errorf("expected context=HelpContextGlobal, got %s", m.Context())
		}
	})

	t.Run("creates model with dashboard context", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, HelpContextDashboard)

		if m.Context() != HelpContextDashboard {
			t.Errorf("expected context=HelpContextDashboard, got %s", m.Context())
		}
	})

	t.Run("creates model with PR detail context", func(t *testing.T) {
		cfg := testConfig()
		m := New(cfg, HelpContextPRDetail)

		if m.Context() != HelpContextPRDetail {
			t.Errorf("expected context=HelpContextPRDetail, got %s", m.Context())
		}
	})

	t.Run("creates model with all context types", func(t *testing.T) {
		cfg := testConfig()
		contexts := []HelpContext{
			HelpContextGlobal,
			HelpContextDashboard,
			HelpContextPRDetail,
			HelpContextDiffView,
			HelpContextCreatePR,
			HelpContextReview,
			HelpContextComposer,
		}

		for _, ctx := range contexts {
			m := New(cfg, ctx)
			if m.Context() != ctx {
				t.Errorf("expected context=%s, got %s", ctx, m.Context())
			}
		}
	})
}

func TestInit(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)

	cmd := m.Init()

	// Init should return nil (no startup commands needed)
	if cmd != nil {
		t.Error("expected Init() to return nil")
	}
}

func TestSetSize(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)

	m.SetSize(100, 50)

	// Verify dimensions are stored (indirectly via view rendering)
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after SetSize")
	}
}

func TestContext(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name    string
		context HelpContext
	}{
		{"global", HelpContextGlobal},
		{"dashboard", HelpContextDashboard},
		{"prdetail", HelpContextPRDetail},
		{"diffview", HelpContextDiffView},
		{"createpr", HelpContextCreatePR},
		{"review", HelpContextReview},
		{"composer", HelpContextComposer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(cfg, tt.context)
			if m.Context() != tt.context {
				t.Errorf("expected %s, got %s", tt.context, m.Context())
			}
		})
	}
}

// =============================================================================
// Update Tests - Key Handling
// =============================================================================

func TestUpdate_EscClosesHelp(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected a command for Esc key")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(CloseHelpMsg); !ok {
		t.Errorf("expected CloseHelpMsg, got %T", resultMsg)
	}
}

func TestUpdate_QClosesHelp(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected a command for q key")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(CloseHelpMsg); !ok {
		t.Errorf("expected CloseHelpMsg, got %T", resultMsg)
	}
}

func TestUpdate_QuestionMarkClosesHelp(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected a command for ? key")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(CloseHelpMsg); !ok {
		t.Errorf("expected CloseHelpMsg, got %T", resultMsg)
	}
}

func TestUpdate_JScrollsDown(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	// Get initial scroll position
	initialYOffset := m.viewport.YOffset

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Scroll position should have changed (or stayed at 0 if content fits)
	// This verifies the key is handled
	if m.viewport.YOffset < initialYOffset {
		t.Error("expected scroll position to increase or stay same")
	}
}

func TestUpdate_KScrollsUp(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	// First scroll down
	m.viewport.LineDown(5)
	initialYOffset := m.viewport.YOffset

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Scroll position should have decreased
	if m.viewport.YOffset > initialYOffset {
		t.Error("expected scroll position to decrease or stay same")
	}
}

func TestUpdate_DownArrowScrollsDown(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	_, cmd := m.Update(msg)

	// Should not produce a close command
	if cmd != nil {
		resultMsg := cmd()
		if _, ok := resultMsg.(CloseHelpMsg); ok {
			t.Error("down arrow should not close help")
		}
	}
}

func TestUpdate_UpArrowScrollsUp(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyUp}
	_, cmd := m.Update(msg)

	// Should not produce a close command
	if cmd != nil {
		resultMsg := cmd()
		if _, ok := resultMsg.(CloseHelpMsg); ok {
			t.Error("up arrow should not close help")
		}
	}
}

func TestUpdate_PageDownScrollsPage(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyPgDown}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Verify key was handled (model updated without error)
	if m.Context() != HelpContextGlobal {
		t.Error("model context should be preserved")
	}
}

func TestUpdate_PageUpScrollsPage(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyPgUp}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Verify key was handled (model updated without error)
	if m.Context() != HelpContextGlobal {
		t.Error("model context should be preserved")
	}
}

func TestUpdate_HomeGoesToTop(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	// Scroll down first
	m.viewport.LineDown(10)

	msg := tea.KeyMsg{Type: tea.KeyHome}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.viewport.YOffset != 0 {
		t.Errorf("expected YOffset=0 after Home, got %d", m.viewport.YOffset)
	}
}

func TestUpdate_EndGoesToBottom(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyEnd}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Should be at or near bottom (exact position depends on content)
	// Just verify the key is handled
	if m.Context() != HelpContextGlobal {
		t.Error("model context should be preserved")
	}
}

func TestUpdate_GGoesToTop(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	// Scroll down first
	m.viewport.LineDown(10)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.viewport.YOffset != 0 {
		t.Errorf("expected YOffset=0 after 'g', got %d", m.viewport.YOffset)
	}
}

func TestUpdate_ShiftGGoesToBottom(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Just verify the key is handled
	if m.Context() != HelpContextGlobal {
		t.Error("model context should be preserved")
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Verify dimensions are updated
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after resize")
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestView_NotEmpty(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestView_ContainsTitle(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	view := m.View()

	if !strings.Contains(view, "Help") {
		t.Error("expected view to contain 'Help' title")
	}
}

func TestView_ContainsGlobalKeys(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	view := m.View()

	// Should contain global key information
	globalKeys := []string{"GLOBAL", "quit", "help"}
	for _, key := range globalKeys {
		if !strings.Contains(strings.ToLower(view), strings.ToLower(key)) {
			t.Errorf("expected view to contain %q", key)
		}
	}
}

func TestView_ContainsNavigationKeys(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	view := m.View()

	// Should contain navigation information
	navInfo := []string{"NAVIGATION", "down", "up"}
	for _, info := range navInfo {
		if !strings.Contains(strings.ToLower(view), strings.ToLower(info)) {
			t.Errorf("expected view to contain %q", info)
		}
	}
}

func TestView_DashboardContext(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextDashboard)
	m.SetSize(80, 60) // Larger height to fit all content

	// Check the generated content (not truncated by viewport)
	content := m.generateHelpContent()

	// Should contain dashboard-specific keys
	dashboardKeys := []string{"refresh", "search", "create"}
	for _, key := range dashboardKeys {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(key)) {
			t.Errorf("expected dashboard content to contain %q", key)
		}
	}
}

func TestView_PRDetailContext(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextPRDetail)
	m.SetSize(80, 60)

	// Check the generated content (not truncated by viewport)
	content := m.generateHelpContent()

	// Should contain PR detail-specific keys
	prDetailKeys := []string{"diff", "merge"}
	for _, key := range prDetailKeys {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(key)) {
			t.Errorf("expected PR detail content to contain %q", key)
		}
	}
}

func TestView_DiffViewContext(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextDiffView)
	m.SetSize(80, 60)

	// Check the generated content (not truncated by viewport)
	content := m.generateHelpContent()

	// Should contain diff view-specific keys
	diffKeys := []string{"next", "prev"}
	for _, key := range diffKeys {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(key)) {
			t.Errorf("expected diff view content to contain %q", key)
		}
	}
}

func TestView_ReviewContext(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextReview)
	m.SetSize(80, 60)

	// Check the generated content (not truncated by viewport)
	content := m.generateHelpContent()

	// Should contain review-specific keys
	reviewKeys := []string{"approve", "submit"}
	for _, key := range reviewKeys {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(key)) {
			t.Errorf("expected review content to contain %q", key)
		}
	}
}

func TestView_ComposerContext(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextComposer)
	m.SetSize(80, 60)

	// Check the generated content (not truncated by viewport)
	content := m.generateHelpContent()

	// Should contain composer-specific keys
	composerKeys := []string{"generate", "accept"}
	for _, key := range composerKeys {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(key)) {
			t.Errorf("expected composer content to contain %q", key)
		}
	}
}

func TestView_ContainsFooterHints(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)
	m.SetSize(80, 24)

	view := m.View()

	// Should contain footer hints
	if !strings.Contains(strings.ToLower(view), "scroll") || !strings.Contains(strings.ToLower(view), "close") {
		t.Error("expected view to contain scroll and close hints")
	}
}

// =============================================================================
// Help Content Generation Tests
// =============================================================================

func TestGenerateHelpContent_NotEmpty(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)

	content := m.generateHelpContent()

	if content == "" {
		t.Error("expected non-empty help content")
	}
}

func TestGenerateHelpContent_AllContexts(t *testing.T) {
	cfg := testConfig()
	contexts := []HelpContext{
		HelpContextGlobal,
		HelpContextDashboard,
		HelpContextPRDetail,
		HelpContextDiffView,
		HelpContextCreatePR,
		HelpContextReview,
		HelpContextComposer,
	}

	for _, ctx := range contexts {
		t.Run(ctx.String(), func(t *testing.T) {
			m := New(cfg, ctx)
			content := m.generateHelpContent()

			if content == "" {
				t.Errorf("expected non-empty content for context %s", ctx)
			}

			// All contexts should have global section
			if !strings.Contains(content, "GLOBAL") {
				t.Errorf("expected GLOBAL section for context %s", ctx)
			}

			// All contexts should have navigation section
			if !strings.Contains(content, "NAVIGATION") {
				t.Errorf("expected NAVIGATION section for context %s", ctx)
			}
		})
	}
}

// =============================================================================
// KeyBinding Tests
// =============================================================================

func TestKeyBinding_Structure(t *testing.T) {
	kb := KeyBinding{
		Key:         "q",
		Description: "Quit application",
	}

	if kb.Key != "q" {
		t.Errorf("expected Key='q', got %q", kb.Key)
	}

	if kb.Description != "Quit application" {
		t.Errorf("expected Description='Quit application', got %q", kb.Description)
	}
}

func TestGetContextBindings_Global(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextGlobal)

	bindings := m.getContextBindings()

	// Global context should return empty since it's shown for all contexts
	// The global bindings are shown separately
	if bindings == nil {
		t.Error("expected non-nil bindings slice")
	}
}

func TestGetContextBindings_Dashboard(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextDashboard)

	bindings := m.getContextBindings()

	if len(bindings) == 0 {
		t.Error("expected dashboard bindings")
	}

	// Check for expected dashboard keys
	found := false
	for _, b := range bindings {
		if strings.Contains(strings.ToLower(b.Description), "refresh") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected refresh binding in dashboard context")
	}
}

func TestGetContextBindings_PRDetail(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextPRDetail)

	bindings := m.getContextBindings()

	if len(bindings) == 0 {
		t.Error("expected PR detail bindings")
	}

	// Check for expected PR detail keys
	found := false
	for _, b := range bindings {
		if strings.Contains(strings.ToLower(b.Description), "diff") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected view diff binding in PR detail context")
	}
}

// =============================================================================
// Message Tests
// =============================================================================

func TestCloseHelpMsg(t *testing.T) {
	msg := CloseHelpMsg{}
	// Just verify it's a valid type (used for type assertion)
	_ = msg
}

func TestBackMsg(t *testing.T) {
	msg := BackMsg{}
	// Just verify it's a valid type (used for type assertion)
	_ = msg
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestFullFlow(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, HelpContextDashboard)

	// Initialize
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected no init command")
	}

	// Set size
	m.SetSize(80, 24)

	// Verify view
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Scroll down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Close with Esc
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd = m.Update(msg)

	if cmd == nil {
		t.Fatal("expected close command")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(CloseHelpMsg); !ok {
		t.Errorf("expected CloseHelpMsg, got %T", resultMsg)
	}
}
