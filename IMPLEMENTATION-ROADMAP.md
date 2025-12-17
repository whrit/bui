# Implementation Roadmap

## Current Status Summary

The codebase has a solid foundation with Phase 1 (Core Infrastructure) now complete:

| Component           | Status      | Coverage | Notes                                                   |
|---------------------|-------------|----------|--------------------------------------------------------|
| Configuration       | ✅ Complete | -        | TOML-based, auto-generated, all settings defined        |
| LLM Subsystem       | ✅ Complete | -        | 3 providers (OpenAI, Anthropic, Ollama), full streaming |
| LLM Core Types      | ✅ Complete | -        | Refactored to `internal/llm/core` to fix import cycle   |
| Syntax Highlighting | ✅ Complete | -        | Chroma-based, diff-aware, theme support                 |
| UI Theming          | ✅ Complete | -        | Palette, accent colors, Lip Gloss styles                |
| Entry Point         | ✅ Complete | -        | Config loading, Bubble Tea initialization               |
| **internal/gh/**    | ✅ Complete | 89.9%    | GitHub CLI wrapper with full PR lifecycle support       |
| **internal/git/**   | ✅ Complete | 89.7%    | Git utilities for repo, commits, diffs, branches        |
| **UI Keymap**       | ✅ Complete | 82.9%    | 28 key bindings, config integration, help generation    |
| **UI Components**   | ✅ Complete | 82.9%    | Spinner, StatusBadge, CheckIndicator, ErrorDisplay, ConfirmDialog |

## Phase 1 Implementation Details

### internal/gh/ - GitHub CLI Wrapper (✅ Complete)

```
internal/gh/
├── types.go        # PR, Review, Check, Label, Reviewer, PRFilters, MergeMethod, ReviewEvent
├── client.go       # Client struct with CommandExecutor interface for DI
├── pr.go           # ListPRs, GetPR, GetPRDiff, CreatePR, MergePR, ClosePR
├── review.go       # GetReviews, SubmitReview
├── checks.go       # GetChecks
├── client_test.go  # Client construction and concurrency tests
├── pr_test.go      # PR operation tests with mock executor
├── review_test.go  # Review operation tests
└── checks_test.go  # Check operation tests
```

**Functions implemented:**
- `ListPRs(filters PRFilters) → ([]PR, error)`
- `GetPR(number int) → (*PR, error)`
- `GetPRDiff(number int) → (string, error)`
- `CreatePR(base, head, title, body string, isDraft bool) → (*PR, error)`
- `MergePR(number int, method MergeMethod) → error`
- `ClosePR(number int) → error`
- `GetReviews(prNumber int) → ([]Review, error)`
- `SubmitReview(prNumber int, body string, event ReviewEvent) → error`
- `GetChecks(prNumber int) → ([]Check, error)`

### internal/git/ - Git Utilities (✅ Complete)

```
internal/git/
├── executor.go     # CommandExecutor interface for dependency injection
├── repo.go         # GetRepoInfo, GetRootPath, IsInsideWorkTree, GetRemoteURL, ListRemotes
├── commits.go      # GetCurrentBranch, GetCommitLog, GetCommitMessages, GetCommit, GetMergeBase
├── diff.go         # GetDiff, GetDiffStat, GetFileDiffs, GetStagedDiff, GetUnstagedDiff
├── branch.go       # ListBranches, GetBranch, BranchExists, GetDefaultBranch, CreateBranch, DeleteBranch
├── repo_test.go    # Repository operation tests
├── commits_test.go # Commit operation tests
├── diff_test.go    # Diff operation tests
└── branch_test.go  # Branch operation tests
```

**Functions implemented:**
- Repo: `GetRepoInfo()`, `GetRootPath()`, `IsInsideWorkTree()`, `GetRemoteURL()`, `ListRemotes()`
- Commits: `GetCurrentBranch()`, `GetCommitLog()`, `GetCommitMessages()`, `GetCommit()`, `GetCommitCount()`, `GetMergeBase()`, `GetHeadCommit()`, `ResolveRef()`
- Diffs: `GetDiff()`, `GetDiffStat()`, `GetFileDiffs()`, `GetStagedDiff()`, `GetUnstagedDiff()`, `GetDiffForFile()`, `GetChangedFiles()`, `HasChanges()`, `HasStagedChanges()`, `HasUnstagedChanges()`
- Branches: `ListBranches()`, `ListLocalBranches()`, `ListRemoteBranches()`, `GetBranch()`, `BranchExists()`, `GetUpstreamBranch()`, `GetDefaultBranch()`, `SetUpstream()`, `CreateBranch()`, `DeleteBranch()`, `RenameBranch()`, `CheckoutBranch()`, `GetBranchCommitsBehindAhead()`

### internal/ui/keymap.go - Keyboard Bindings (✅ Complete)

**28 key bindings organized by category:**
- Global: Quit, Help, Confirm, Cancel
- Navigation: Up, Down, Left, Right, PageUp, PageDown, Home, End
- Vim-style: NextItem (j), PrevItem (k)
- Actions: Select, Back, Refresh, Search, Filter
- PR-specific: OpenInBrowser, ViewDiff, ViewDetails, CreatePR, MergePR, ClosePR, ApprovePR, RequestChanges
- LLM: GenerateWithLLM, RegenerateLLM, AcceptLLM, EditLLM
- Editor: Submit, ClearAll

**Features:**
- `NewKeyMap(cfg)` - Creates KeyMap with config overrides
- `DefaultKeyMap()` - Returns defaults without config
- `ShortHelp()` / `FullHelp()` - Implements `help.KeyMap` interface
- Config-aware (quit, help, confirm are user-configurable)

### internal/ui/components.go - Reusable Components (✅ Complete)

**Components:**
1. `StatusBadge` - PR state badges (open/closed/merged/draft) with colored backgrounds
2. `CheckIndicator` - CI check status with icons (✓, ✗, ○, ◐, ⊘, ⊗)
3. `Spinner` - Loading spinner with optional message (uses bubbles/spinner)
4. `ErrorDisplay` - Bordered error box with optional title
5. `ConfirmDialog` - Modal confirmation dialog with keyboard navigation

**Helpers:**
- `Truncate(s string, maxLen int)` - Truncate with ellipsis
- `PadRight(s string, width int)` - Fixed-width padding
- `RelativeTime(t time.Time)` - Human-readable time ("2 hours ago")
- `Divider(width int, p Palette)` - Horizontal divider line

---

## Remaining Work (Required for v1)

| Component          | Status         | Priority      |
|--------------------|----------------|---------------|
| Dashboard Screen   | ❌ Not started | P0 - Critical |
| PR Detail Screen   | ❌ Not started | P0 - Critical |
| Diff Viewer Screen | ❌ Not started | P0 - Critical |
| Composer Screen    | ❌ Not started | P1 - High     |
| Create PR Wizard   | ❌ Not started | P1 - High     |
| Review Screen      | ❌ Not started | P1 - High     |
| UI Layout          | ❌ Not started | P1 - High     |

---

## Phase 2: Main Screens (Next)

### 5. Dashboard Screen (internal/app/dashboard/)
- PR list with status indicators
- Keyboard navigation (j/k, enter to select)
- Filter/search capability
- Auto-refresh toggle
- Actions: open detail, create new PR

### 6. PR Detail Screen (internal/app/prdetail/)
- Display PR metadata (title, author, labels, reviewers)
- Show description (markdown rendered)
- CI checks status
- Merge controls
- LLM "Summarize" action (wired to existing provider)
- Navigation to diff view

### 7. Diff Viewer Screen (internal/app/diffview/)
- Scrollable viewport with syntax-highlighted diff (use existing syntax.HighlightDiff)
- File tree/navigation
- Line numbers toggle
- Hunk selection for LLM context
- Navigation back to PR detail

---

## Phase 3: LLM-Integrated Screens

### 8. Composer Screen (internal/app/composer/)
- Multiline text editor (use bubbles/textarea)
- LLM streaming display
- Edit/accept/regenerate controls
- Cancel LLM generation

### 9. Create PR Wizard (internal/app/createpr/)
- Step 1: Base/head branch selection
- Step 2: LLM draft generation (uses prompts/pr_create)
- Step 3: Edit title/body in composer
- Step 4: Submit via gh pr create

### 10. Review Screen (internal/app/review/)
- Review type selection (approve/request changes/comment)
- LLM-assisted comment drafting
- Preview and submit

---

## Phase 4: Polish & Integration

### 11. Root Model Enhancement (internal/app/root/)
- Screen state machine
- Navigation between screens
- Global keybindings (quit, help)
- Error toast/notifications

### 12. Help Screen
- Generated from keymap
- Context-sensitive

### 13. internal/ui/layout.go
- Header component
- Footer/status bar
- Screen frame

---

## Architecture Notes

### Design Patterns Used
1. **Dependency Injection**: `CommandExecutor` interface enables testing without actual CLI calls
2. **Factory Pattern**: LLM providers selected via factory based on config
3. **Type Aliases**: `internal/llm/types.go` re-exports from `internal/llm/core` to avoid import cycles
4. **Table-Driven Tests**: All tests use subtests with `t.Run()`
5. **Error Wrapping**: Proper `fmt.Errorf("%w")` for error chains

### Test Coverage Summary
- `internal/gh/`: 89.9%
- `internal/git/`: 89.7%
- `internal/ui/`: 82.9%
- All tests pass with `-race` detector
- All linter checks pass (golangci-lint)

### File Statistics
- Total Go source files: ~30
- Total lines of code: ~7,500+
- Test files: 12
