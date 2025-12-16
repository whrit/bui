Below is a **ready-to-commit `PRD.md`**.
It is clean Markdown, repo-ready, and formatted so you can drop it directly into `docs/PRD.md` or the repo root without edits.

---

```md
# PRD: Terminal-Native GitHub Pull Request Manager with Integrated LLM Copilot

**Working name:** bui  
**Platform:** macOS (primary), Linux (secondary), Windows (best-effort)  
**Distribution:** Single self-contained binary  
**Core Stack:** Go · Bubble Tea · Bubbles · Lip Gloss  

---

## 1. Executive Summary

`bui` is a **full-lifecycle GitHub Pull Request client** built as a modern, beautiful terminal user interface (TUI). It enables developers to **create, review, manage, and merge pull requests** without leaving the terminal, while integrating **first-class LLM assistance**, **syntax-highlighted diffs**, and a **mandatory configuration system**.

Unlike dashboards or thin wrappers around `gh`, `bui` is a **stateful, opinionated, daily-driver application** designed for speed, clarity, and trust.

---

## 2. Core Principles (Non-Negotiable)

1. **LLM-Native**  
   AI assistance is built into primary workflows and is never an afterthought.

2. **Readable Code First**  
   Syntax highlighting and diff legibility are first-class UX requirements.

3. **Configuration Is Mandatory**  
   All behavior is explicit, inspectable, and configurable from day one.

4. **Terminal-First Excellence**  
   Fully keyboard-driven, low latency, predictable behavior.

5. **User Agency Always**  
   LLMs propose; users decide. No silent or automatic mutations.

---

## 3. Problem Statement

GitHub PR workflows today are fragmented:

- Browsers cause context switching and latency
- IDE integrations are heavy and inconsistent
- Existing TUIs are either read-only or visually unpolished
- LLM assistance is bolted on, opaque, or browser-only

There is no **beautiful, terminal-native PR client** that:
- covers the entire PR lifecycle,
- integrates AI transparently,
- and remains fast, inspectable, and trustworthy.

---

## 4. Goals & Non-Goals

### Goals
- Full PR lifecycle support (create → review → merge)
- Integrated, transparent LLM copilot
- Syntax-highlighted diffs comparable to GitHub web
- Modern, calm, aesthetic TUI
- Single-binary installation

### Non-Goals
- Replacing `gh` as the GitHub integration layer
- Managing GitHub Issues/Projects in v1
- Autonomous or background AI actions

---

## 5. Target Users

**Primary**
- Senior engineers
- Open-source maintainers
- Terminal-native macOS users

**Secondary**
- DevOps / infra engineers
- Consultants working across many repositories
- Users of `gh`, `lazygit`, or `tig`

---

## 6. Technology Stack

### Language & UI
- Go
- Bubble Tea (state machine / event loop)
- Bubbles (lists, inputs, viewports)
- Lip Gloss (styling, layout, theming)

### GitHub & Git
- GitHub CLI (`gh`) — required dependency
- Native `git` shell calls for local context

### LLMs (Core Feature)
- Provider-agnostic abstraction
- Required providers at launch:
  - OpenAI
  - Anthropic
  - Local (Ollama or compatible HTTP API)

### Syntax Highlighting (Core Feature)
- Diff-aware syntax highlighting
- Language detection per file
- Theme-aware rendering

### Configuration (Core Feature)
- TOML-based configuration
- Auto-generated on first run
- Required for application startup

---

## 7. High-Level Architecture

```

┌──────────────────────────────────────────┐
│              Bubble Tea App               │
│                                          │
│  ┌──────── Screen Models ─────────────┐  │
│  │ Dashboard | PR | Diff | Composer   │  │
│  │ Create PR | Review | Help          │  │
│  └───────────────┬───────────────────┘  │
│                  │ messages / commands   │
│  ┌───────────────▼───────────────────┐  │
│  │        Core Application Services   │  │
│  │                                   │  │
│  │  gh driver | git | llm | config   │  │
│  │  syntax highlighter | theme       │  │
│  └───────────────┬───────────────────┘  │
│                  │ shell / HTTP calls    │
│  ┌───────────────▼───────────────────┐  │
│  │          External Systems          │  │
│  │ gh CLI | git | LLM APIs | FS       │  │
│  └───────────────────────────────────┘  │
└──────────────────────────────────────────┘

````

**Key invariant:**  
UI layers never talk directly to GitHub APIs — all GitHub interaction flows through `gh`.

---

## 8. Repository Structure

```txt
cmd/bui/main.go

internal/
  app/
    root/
    dashboard/
    prdetail/
    diffview/
    composer/
    createpr/
    review/

  gh/                  # GitHub CLI wrapper (JSON-first)
  git/                 # git log, diff, branch utilities

  llm/                 # REQUIRED LLM subsystem
    providers/
      openai/
      anthropic/
      local/
    prompts/
    streaming.go

  syntax/              # REQUIRED syntax highlighting
    diff.go
    languages.go
    themes.go

  config/              # REQUIRED config system
    load.go
    defaults.go
    schema.go

  ui/
    theme.go
    keymap.go
    layout.go
    components.go
````

---

## 9. Configuration System (Mandatory)

### First-Run Behavior

* If no config exists, generate one automatically
* Notify user of config path
* Application will not start silently without config

### Config Location

* macOS/Linux: `~/.config/bui/config.toml`
* Windows: `%APPDATA%\bui\config.toml`

### Example Config

```toml
[ui]
theme = "dark"
accent = "indigo"
syntax_theme = "github-dark"
show_line_numbers = true

[keys]
quit = "q"
help = "?"
confirm = "enter"

[llm]
provider = "openai"
model = "gpt-4.1"
max_tokens = 800
temperature = 0.2
stream = true

[llm.privacy]
send_full_diff = false
max_diff_lines = 300

[git]
default_base = "main"

[behavior]
confirm_merge = true
auto_refresh = true
```

---

## 10. LLM System (Core)

### Philosophy

* LLMs are explicit, assistive copilots
* All invocations are user-initiated
* Output is always previewed and editable
* Streaming output is required

### Core Use Cases

1. Draft PR title and body
2. Summarize PR changes
3. Draft review comments
4. Draft request-changes explanations

### Interface

```go
type Provider interface {
  Stream(prompt Prompt) (<-chan Token, error)
  Name() string
}
```

### Inputs

* Commit messages
* Diff stats
* Selected hunks
* Config-defined limits

### Safety Controls

* Token limits
* Line limits
* Explicit diff-sharing toggles

---

## 11. Syntax Highlighting (Core)

### Requirements

* Unified diff highlighting
* Language-aware syntax coloring
* Theme-aware rendering
* Performant on large diffs

### Rendering Pipeline

```
raw diff
 → parse hunks
 → detect language
 → syntax highlight
 → apply theme
 → render viewport
```

---

## 12. Screens & UX

### Dashboard

* PR list with status indicators
* Keyboard navigation
* Saved filters

### PR Detail View

* Metadata, description, checks
* Merge actions
* LLM “Summarize” action

### Diff Viewer

* Syntax-highlighted diff
* Scrollable viewport
* Optional line numbers
* Selectable hunks for LLM context

### Composer

* Multiline editor
* Live LLM streaming
* Manual edits always allowed

### Create PR Wizard

1. Base/head selection
2. LLM draft (entry point, skippable)
3. Edit
4. Submit via `gh pr create`

---

## 13. Async & Concurrency Model

* All external operations use Bubble Tea commands
* Long operations show spinners and progress
* LLM streaming via goroutines + `Program.Send`
* UI remains responsive at all times

---

## 14. Styling & Visual Identity

### Design Goals

* Calm, modern, minimal
* Subtle borders
* Clear focus and selection states
* No ASCII clutter

### Theming

* Config-driven UI themes
* Syntax theme matches UI theme
* Accent color consistent across components

---

## 15. Error Handling

* Inline non-fatal errors
* Full diagnostics for fatal errors
* Common cases:

  * Missing `gh` auth
  * Repository mismatch
  * LLM provider unavailable
  * Diff size limits exceeded

---

## 16. Security & Privacy

* No GitHub tokens stored
* No implicit LLM calls
* Configurable diff sharing limits
* Explicit disclosure before sending data to LLMs

---

## 17. Roadmap

### v1 (Foundational)

* All core screens
* LLM drafting
* Syntax-highlighted diffs
* Full configuration system

### v2

* Inline review comments
* Saved queries
* Multi-repo dashboard

### v3

* Release notes generation
* PR template awareness
* Plugin hooks

---

## 18. Success Criteria

* Users can go days without opening GitHub.com
* LLM output is consistently helpful and trusted
* Diff readability matches or exceeds GitHub web
* Single-binary install with near-zero friction

---

## 19. Final Positioning

`bui` is not:

* a thin wrapper around `gh`
* a Bubble Tea demo
* a novelty AI tool

It **is**:

* a serious, terminal-native PR client
* opinionated, fast, and beautiful
* AI-augmented without sacrificing control

```
