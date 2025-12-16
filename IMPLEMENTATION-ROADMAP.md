⏺ Implementation Roadmap

Current Status Summary

The codebase has a solid foundation with these core systems complete:

| Component           | Status      | Notes                                                   |
|---------------------|-------------|---------------------------------------------------------|
| Configuration       | ✅ Complete | TOML-based, auto-generated, all settings defined        |
| LLM Subsystem       | ✅ Complete | 3 providers (OpenAI, Anthropic, Ollama), full streaming |
| Syntax Highlighting | ✅ Complete | Chroma-based, diff-aware, theme support                 |
| UI Theming          | ✅ Complete | Palette, accent colors, Lip Gloss styles                |
| Entry Point         | ✅ Complete | Config loading, Bubble Tea initialization               |

Missing Components (Required for v1)

| Component          | Status         | Priority      |
|--------------------|----------------|---------------|
| internal/gh/       | ❌ Not started | P0 - Critical |
| internal/git/      | ❌ Not started | P0 - Critical |
| Dashboard Screen   | ❌ Not started | P0 - Critical |
| PR Detail Screen   | ❌ Not started | P0 - Critical |
| Diff Viewer Screen | ❌ Not started | P0 - Critical |
| Composer Screen    | ❌ Not started | P1 - High     |
| Create PR Wizard   | ❌ Not started | P1 - High     |
| Review Screen      | ❌ Not started | P1 - High     |
| UI Keymap          | ❌ Not started | P0 - Critical |
| UI Layout          | ❌ Not started | P1 - High     |
| UI Components      | ❌ Not started | P0 - Critical |

---
Recommended Implementation Order

Phase 1: Core Infrastructure (Do First)

These are blocking dependencies for all UI screens.

1. internal/gh/ - GitHub CLI Wrapper
internal/gh/
├── client.go       # Main gh wrapper struct
├── pr.go           # PR operations (list, get, create, merge)
├── review.go       # Review operations
├── checks.go       # CI/status checks
└── types.go        # Data structures for PR, Review, Check, etc.

Key functions needed:
- ListPRs(filters) → []PR
- GetPR(number) → PR
- GetPRDiff(number) → string
- CreatePR(base, head, title, body) → PR
- MergePR(number, method) → error
- SubmitReview(number, body, event) → error
- GetChecks(number) → []Check

2. internal/git/ - Git Utilities
internal/git/
├── repo.go         # Repository info, current branch
├── commits.go      # Commit log, messages
├── diff.go         # Generate diffs
└── branch.go       # Branch operations

Key functions needed:
- GetCurrentBranch() → string
- GetCommitLog(base, head) → []Commit
- GetDiffStat(base, head) → DiffStat
- GetDiff(base, head) → string
- GetRepoInfo() → RepoInfo (owner, name)

3. internal/ui/keymap.go - Keyboard Bindings
- Define KeyMap struct with all app keybindings
- Load from config with fallback defaults
- Help text generation for help screen

4. internal/ui/components.go - Reusable Components
- Spinner component
- Status badge (open/closed/merged)
- Check status indicator
- Error display
- Confirmation dialog

---
Phase 2: Main Screens (Build in Order)

5. Dashboard Screen (internal/app/dashboard/)
- PR list with status indicators
- Keyboard navigation (j/k, enter to select)
- Filter/search capability
- Auto-refresh toggle
- Actions: open detail, create new PR

6. PR Detail Screen (internal/app/prdetail/)
- Display PR metadata (title, author, labels, reviewers)
- Show description (markdown rendered)
- CI checks status
- Merge controls
- LLM "Summarize" action (wired to existing provider)
- Navigation to diff view

7. Diff Viewer Screen (internal/app/diffview/)
- Scrollable viewport with syntax-highlighted diff (use existing syntax.HighlightDiff)
- File tree/navigation
- Line numbers toggle
- Hunk selection for LLM context
- Navigation back to PR detail

---
Phase 3: LLM-Integrated Screens

8. Composer Screen (internal/app/composer/)
- Multiline text editor (use bubbles/textarea)
- LLM streaming display
- Edit/accept/regenerate controls
- Cancel LLM generation

9. Create PR Wizard (internal/app/createpr/)
- Step 1: Base/head branch selection
- Step 2: LLM draft generation (uses prompts/pr_create)
- Step 3: Edit title/body in composer
- Step 4: Submit via gh pr create

10. Review Screen (internal/app/review/)
- Review type selection (approve/request changes/comment)
- LLM-assisted comment drafting
- Preview and submit

---
Phase 4: Polish & Integration

11. Root Model Enhancement (internal/app/root/)
- Screen state machine
- Navigation between screens
- Global keybindings (quit, help)
- Error toast/notifications

12. Help Screen
- Generated from keymap
- Context-sensitive

13. internal/ui/layout.go
- Header component
- Footer/status bar
- Screen frame

---
