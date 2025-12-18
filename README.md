# bui

A terminal-native GitHub Pull Request manager with **built-in LLM copilot**, **syntax-highlighted diffs**, and **full PR lifecycle support**.

## Features

- **Dashboard** - Browse and filter PRs with real-time status
- **PR Details** - View metadata, CI checks, reviews, and descriptions
- **Diff Viewer** - Syntax-highlighted diffs with file tree navigation, hunk selection, and search
- **Create PR** - Guided wizard with LLM-powered title/description generation
- **Review** - Submit approvals, request changes, or comments with AI assistance
- **LLM Integration** - OpenAI, Anthropic, or local (Ollama) providers for AI-assisted workflows

## Requirements

- **Go 1.22+**
- **GitHub CLI (`gh`)** - installed and authenticated (`gh auth login`)
- **Git** - for repository operations

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/bui.git
cd bui

# Install dependencies
go mod download

# Build the binary
go build -o bui ./cmd/bui

# Or run directly
go run ./cmd/bui
```

## Quick Start

```bash
# Navigate to any git repository with GitHub remote
cd /path/to/your/repo

# Run bui
./bui
```

On first run, bui will create a configuration file at:
- **macOS/Linux**: `~/.config/bui/config.toml`
- **Windows**: `%APPDATA%\bui\config.toml`

## LLM Setup

bui supports three LLM providers for AI-assisted features:

### OpenAI (Default)

```bash
export OPENAI_API_KEY="your-api-key"
```

### Anthropic

```bash
export ANTHROPIC_API_KEY="your-api-key"
```

Then update your config:
```toml
[llm]
provider = "anthropic"
model = "claude-3-sonnet-20240229"
```

### Local (Ollama)

1. Install [Ollama](https://ollama.ai)
2. Pull a model: `ollama pull mistral`
3. Update your config:

```toml
[llm]
provider = "local"
model = "mistral"

[llm.local]
base_url = "http://localhost:11434"
```

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `?` | Show help |
| `Esc` | Back / Cancel |
| `Enter` | Confirm / Select |

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Next item |
| `k` / `↑` | Previous item |
| `PgDn` | Page down |
| `PgUp` | Page up |
| `Home` | Go to start |
| `End` | Go to end |

### Dashboard

| Key | Action |
|-----|--------|
| `Enter` / `v` | View PR details |
| `c` | Create new PR |
| `r` | Refresh list |
| `/` | Search PRs |
| `f` | Filter by status |
| `o` | Open in browser |

### PR Detail

| Key | Action |
|-----|--------|
| `d` | View diff |
| `m` | Merge PR |
| `x` | Close PR |
| `a` | Approve PR |
| `R` | Request changes |
| `w` | Write review |
| `g` | Generate AI summary |
| `o` | Open in browser |
| `Tab` | Switch sections |
| `1-4` | Jump to section |

### Diff Viewer

| Key | Action |
|-----|--------|
| `Tab` | Toggle file tree / diff focus |
| `]` / `[` | Next / previous file |
| `}` / `{` | Next / previous hunk |
| `s` / `Space` | Toggle hunk selection |
| `/` | Search in diff |
| `n` / `N` | Next / previous match |
| `n` (no search) | Toggle line numbers |
| `←` | Focus file tree |

### Create PR / Review

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `g` | Generate with AI |
| `Ctrl+S` | Submit |
| `d` | Toggle draft mode (Create PR) |

### LLM/Composer

| Key | Action |
|-----|--------|
| `g` | Generate with LLM |
| `G` | Regenerate |
| `Enter` | Accept suggestion |
| `e` | Edit suggestion |
| `Ctrl+U` | Clear all |

## Configuration

Full configuration reference (`config.toml`):

```toml
# =============================================================================
# UI Settings
# =============================================================================
[ui]
theme = "dark"              # "dark" or "light"
accent = "indigo"           # indigo, blue, cyan, teal, purple, pink, amber, green
syntax_theme = "github-dark" # Chroma theme for syntax highlighting
show_line_numbers = true

# =============================================================================
# Key Bindings (customizable)
# =============================================================================
[keys]
quit = "q"
help = "?"
confirm = "enter"

# =============================================================================
# LLM Configuration
# =============================================================================
[llm]
provider = "openai"         # openai, anthropic, or local
model = "gpt-4.1-mini"      # Model name
max_tokens = 800            # Max response tokens
temperature = 0.2           # Response randomness (0.0-1.0)
stream = true               # Stream responses
timeout = 60                # Timeout in seconds

# Privacy controls
[llm.privacy]
send_full_diff = false      # Send complete diff to LLM
max_diff_lines = 300        # Truncate diffs to N lines

# Local LLM (Ollama) settings
[llm.local]
base_url = "http://localhost:11434"

# =============================================================================
# Git Settings
# =============================================================================
[git]
default_base = "main"       # Default base branch for PRs

# =============================================================================
# Behavior Settings
# =============================================================================
[behavior]
confirm_merge = true        # Require confirmation before merge
auto_refresh = true         # Auto-refresh PR list
```

## Screens

### Dashboard
The main screen showing all open PRs in the current repository. Use filters to narrow down PRs by status, author, or search query.

### PR Detail
Detailed view of a single PR with four sections:
1. **Info** - Title, author, dates, merge status
2. **Description** - PR body/description
3. **Checks** - CI/CD status checks
4. **Reviews** - Review status and comments

### Diff Viewer
Split-panel view with:
- **Left panel**: File tree with change indicators (A=Added, M=Modified, D=Deleted, R=Renamed)
- **Right panel**: Syntax-highlighted diff with line numbers

Supports binary file detection, hunk-by-hunk navigation, and in-diff search.

### Create PR
Guided wizard for creating PRs:
1. Select base and head branches
2. Optionally generate title/description with LLM
3. Edit title and description
4. Review and submit (optionally as draft)

### Review
Submit PR reviews with:
- **Approve** - Approve the changes
- **Request Changes** - Request modifications
- **Comment** - Leave a general comment

Supports LLM-assisted review generation.

## Architecture

```
bui/
├── cmd/bui/              # Entry point
├── internal/
│   ├── app/              # Screen models (Bubble Tea MVU pattern)
│   │   ├── root/         # Screen router
│   │   ├── dashboard/    # PR list
│   │   ├── prdetail/     # PR details
│   │   ├── diffview/     # Diff viewer
│   │   ├── createpr/     # PR creation wizard
│   │   ├── review/       # Review submission
│   │   ├── help/         # Help screen
│   │   └── composer/     # Text editor with LLM
│   ├── gh/               # GitHub CLI wrapper
│   ├── git/              # Git operations
│   ├── llm/              # LLM provider abstraction
│   │   ├── providers/    # OpenAI, Anthropic, Local
│   │   └── prompts/      # Prompt templates
│   ├── syntax/           # Diff syntax highlighting
│   ├── config/           # TOML configuration
│   └── ui/               # Shared UI components
└── go.mod
```

## Development

```bash
# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Build for all platforms
GOOS=darwin GOARCH=amd64 go build -o bui-darwin-amd64 ./cmd/bui
GOOS=linux GOARCH=amd64 go build -o bui-linux-amd64 ./cmd/bui
GOOS=windows GOARCH=amd64 go build -o bui-windows-amd64.exe ./cmd/bui

# Run linter (if installed)
golangci-lint run
```

## Troubleshooting

### "gh: command not found"
Install the GitHub CLI: https://cli.github.com/

### "not authenticated"
Run `gh auth login` and follow the prompts.

### "not a git repository"
Make sure you're running bui from within a git repository that has a GitHub remote.

### LLM not responding
- Check your API key is set correctly
- Verify network connectivity
- Check the timeout setting in config (default: 60s)
- For local providers, ensure Ollama is running

### Colors look wrong
- Try a different terminal emulator
- Ensure your terminal supports true color (24-bit)
- Try switching theme: `theme = "light"` in config

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.
