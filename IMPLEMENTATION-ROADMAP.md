# Implementation Roadmap

## Current Status Summary

The codebase has Phase 1 (Core Infrastructure) and Phase 2 (Main Screens) complete:

| Component           | Status      | Coverage | Notes                                                   |
|---------------------|-------------|----------|--------------------------------------------------------|
| Configuration       | ✅ Complete | -        | TOML-based, auto-generated, all settings defined        |
| LLM Subsystem       | ✅ Complete | -        | 3 providers (OpenAI, Anthropic, Ollama), full streaming |
| LLM Core Types      | ✅ Complete | -        | Refactored to `internal/llm/core` to fix import cycle   |
| Syntax Highlighting | ✅ Complete | 60.0%    | Chroma-based, diff-aware, theme support                 |
| UI Theming          | ✅ Complete | -        | Palette, accent colors, Lip Gloss styles                |
| Entry Point         | ✅ Complete | -        | Config loading, Bubble Tea initialization               |
| **internal/gh/**    | ✅ Complete | 89.9%    | GitHub CLI wrapper with full PR lifecycle support       |
| **internal/git/**   | ✅ Complete | 89.7%    | Git utilities for repo, commits, diffs, branches        |
| **UI Keymap**       | ✅ Complete | 82.9%    | 28 key bindings, config integration, help generation    |
| **UI Components**   | ✅ Complete | 82.9%    | Spinner, StatusBadge, CheckIndicator, ErrorDisplay, ConfirmDialog |
| **Dashboard Screen**| ✅ Complete | 50.1%    | PR list, navigation, filters, auto-refresh              |
| **PR Detail Screen**| ✅ Complete | 67.6%    | Metadata, checks, reviews, merge, LLM summarize         |
| **Diff Viewer**     | ✅ Complete | 82.0%    | Split panel, syntax highlighting, file tree             |
| **Root Navigation** | ✅ Complete | 100%     | Screen state machine, message routing                   |

---

## Phase 1 Implementation Details (Complete)

### internal/gh/ - GitHub CLI Wrapper (✅ Complete)

```
internal/gh/
├── types.go        # PR, Review, Check, Label, Reviewer, PRFilters, MergeMethod, ReviewEvent
├── client.go       # Client struct with CommandExecutor interface for DI
├── pr.go           # ListPRs, GetPR, GetPRDiff, CreatePR, MergePR, ClosePR
├── review.go       # GetReviews, SubmitReview
├── checks.go       # GetChecks
└── *_test.go       # Comprehensive tests
```

### internal/git/ - Git Utilities (✅ Complete)

```
internal/git/
├── executor.go     # CommandExecutor interface for dependency injection
├── repo.go         # GetRepoInfo, GetRootPath, IsInsideWorkTree, GetRemoteURL, ListRemotes
├── commits.go      # GetCurrentBranch, GetCommitLog, GetCommitMessages, GetCommit, GetMergeBase
├── diff.go         # GetDiff, GetDiffStat, GetFileDiffs, GetStagedDiff, GetUnstagedDiff
├── branch.go       # ListBranches, GetBranch, BranchExists, GetDefaultBranch, CreateBranch, DeleteBranch
└── *_test.go       # Comprehensive tests
```

### internal/ui/ - UI System (✅ Complete)

- **keymap.go**: 28 key bindings with config integration
- **components.go**: StatusBadge, CheckIndicator, Spinner, ErrorDisplay, ConfirmDialog
- **theme.go**: Palette, Styles, accent colors

---

## Phase 2 Implementation Details (Complete)

### Dashboard Screen (✅ Complete)

```
internal/app/dashboard/
├── model.go        # Main Bubble Tea model with State enum
├── view.go         # Header, PR list, footer rendering
├── update.go       # Message handling, key navigation
├── messages.go     # PRsLoadedMsg, OpenPRDetailMsg, CreatePRMsg, etc.
├── delegate.go     # Custom list delegates for PR items
└── model_test.go   # 31 tests
```

**Features:**
- PR list with status badges, author, branch info, +/- stats
- j/k or arrow navigation, enter to select
- / for search/filter mode
- r for refresh, c for create PR
- Auto-refresh toggle (30s interval)
- Loading spinner and error states

### PR Detail Screen (✅ Complete)

```
internal/app/prdetail/
├── model.go        # Main model with State and Section enums
├── view.go         # Header, tab bar, content, footer
├── update.go       # Message handling, navigation
├── messages.go     # Data loading, LLM streaming, action messages
├── sections.go     # Info, Description, Checks, Reviews renderers
└── model_test.go   # 56 tests
```

**Features:**
- Tabbed sections: Info, Description, Checks, Reviews
- Tab/Shift-Tab and number keys (1-4) for section navigation
- PR metadata: title, status badge, author, branches, labels, stats
- CI checks with status indicators and pass/fail counts
- Reviews with state icons (approved, changes requested, etc.)
- Actions: View diff (d), Open browser (o), Merge (m), Close (x), Summarize (g)
- Confirmation dialogs for merge/close
- LLM streaming support for AI summaries
- Scrollable viewport with j/k navigation

### Diff Viewer Screen (✅ Complete)

```
internal/app/diffview/
├── model.go        # Main model with Focus enum
├── view.go         # Split panel layout
├── update.go       # Key handling, panel switching
├── messages.go     # DiffLoadedMsg, BackToPRDetailMsg, etc.
├── filetree.go     # File list with status indicators
├── parser.go       # Unified diff parser
├── parser_test.go  # Parser tests
└── model_test.go   # Model tests
```

**Features:**
- Split panel: 25% file tree, 75% diff content
- File tree with status symbols (A/M/D/R), +/- counts
- Syntax-highlighted diff (uses internal/syntax)
- Optional line numbers (toggle with 'n')
- Tab to switch focus between panels
- j/k scroll, [/] to navigate files
- Esc to go back to PR detail
- Hunk selection infrastructure for LLM context

### Root Model Navigation (✅ Complete)

```
internal/app/root/
├── model.go        # Screen state machine, message routing
└── model_test.go   # Navigation and routing tests
```

**Features:**
- Screen enum: Dashboard, PRDetail, DiffView
- Navigation message handling from all screens
- Window size propagation to active screen
- Global quit handler (configurable key + Ctrl+C)

---

## Remaining Work (Required for v1)

| Component          | Status         | Priority      |
|--------------------|----------------|---------------|
| Composer Screen    | ❌ Not started | P1 - High     |
| Create PR Wizard   | ❌ Not started | P1 - High     |
| Review Screen      | ❌ Not started | P1 - High     |
| UI Layout          | ❌ Not started | P1 - High     |
| Help Screen        | ❌ Not started | P2 - Medium   |

---

## Phase 3: LLM-Integrated Screens (Next)

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
- Error toast/notifications
- Loading overlay

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
6. **Screen State Machine**: Root model manages active screen and routes messages

### Test Coverage Summary
- `internal/gh/`: 89.9%
- `internal/git/`: 89.7%
- `internal/ui/`: 82.9%
- `internal/app/dashboard/`: 50.1%
- `internal/app/prdetail/`: 67.6%
- `internal/app/diffview/`: 82.0%
- `internal/app/root/`: 100%
- All tests pass with `-race` detector
- All linter checks pass (golangci-lint)

### File Statistics
- Total Go source files: 59
- Total lines of code: ~12,000+
- Test files: 20+
