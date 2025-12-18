# bui

A terminal-native GitHub Pull Request manager with **built-in LLM copilot**, **syntax-highlighted diffs**, and **full PR lifecycle support**.

## Features

- **Dashboard** - Browse and filter PRs with real-time status
- **PR Details** - View metadata, CI checks, reviews, and descriptions
- **Diff Viewer** - Syntax-highlighted diffs with file tree navigation, hunk selection, and search
- **Create PR** - Guided wizard with LLM-powered title/description generation
- **Review** - Submit approvals, request changes, or comments with AI assistance
- **LLM Integration** - Access 300+ models via OpenRouter (OpenAI, Anthropic, Llama, Mistral, etc.) or local (Ollama)

## Requirements

- **Go 1.22+**
- **GitHub CLI (`gh`)** - installed and authenticated
- **Git** - for repository operations

## GitHub Authentication

bui uses the [GitHub CLI](https://cli.github.com/) for all GitHub operations. The CLI handles OAuth authentication securely:

```bash
# Install GitHub CLI (if not installed)
# macOS: brew install gh
# Linux: https://github.com/cli/cli/blob/trunk/docs/install_linux.md
# Windows: winget install GitHub.cli

# Authenticate with GitHub (one-time setup)
gh auth login
```

The CLI will guide you through the OAuth flow:
1. Choose GitHub.com or GitHub Enterprise
2. Select HTTPS or SSH for git operations
3. Authenticate via browser (recommended) or paste a token

Your credentials are stored securely by the GitHub CLI - bui never handles tokens directly.

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

bui uses [OpenRouter](https://openrouter.ai) to provide access to 300+ AI models through a single API key.

### Quick Setup

1. **Get an API key** at [openrouter.ai/keys](https://openrouter.ai/keys)

2. **Create a `.env` file** in your project directory:
   ```bash
   cp .env.example .env
   ```

3. **Add your API key:**
   ```bash
   # .env
   OPENROUTER_API_KEY=sk-or-your-api-key-here
   ```

That's it! bui will automatically load the `.env` file on startup.

### Environment Files

bui supports multiple `.env` files with the following priority:
- `.env.local` - Personal overrides (highest priority)
- `.env` - Base configuration

**Note:** Existing environment variables are never overwritten by `.env` files.

### Available Models

OpenRouter provides access to models from multiple providers:

| Provider | Example Models |
|----------|----------------|
| Anthropic | `anthropic/claude-3-5-sonnet`, `anthropic/claude-3-opus` |
| OpenAI | `openai/gpt-4o`, `openai/gpt-4-turbo` |
| Meta | `meta-llama/llama-3.1-70b-instruct` |
| Mistral | `mistralai/mistral-large` |
| Google | `google/gemini-pro-1.5` |

See [openrouter.ai/models](https://openrouter.ai/models) for the full list.

To change models, update your config:
```toml
[llm]
model = "openai/gpt-4o"  # or any OpenRouter model
```

### Local (Ollama)

For offline/local models using Ollama:

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
provider = "openrouter"     # openrouter or local
model = "anthropic/claude-3-5-sonnet"  # OpenRouter model (provider/model format)
max_tokens = 800            # Max response tokens
temperature = 0.2           # Response randomness (0.0-2.0)
stream = true               # Stream responses
timeout = 60                # Timeout in seconds

# Privacy controls
[llm.privacy]
send_full_diff = false      # Send complete diff to LLM
max_diff_lines = 300        # Truncate diffs to N lines

# OpenRouter settings (optional)
[llm.openrouter]
# base_url = "https://openrouter.ai/api/v1"  # Override API endpoint
site_url = ""               # Your app URL (for OpenRouter rankings)
site_name = "bui"           # Your app name (for OpenRouter rankings)

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
│   │   ├── providers/    # OpenRouter, Local (Ollama)
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
- Check your `OPENROUTER_API_KEY` is set in `.env` or environment
- Verify network connectivity to openrouter.ai
- Check the timeout setting in config (default: 60s)
- For local providers, ensure Ollama is running (`ollama serve`)

### ".env file not loading"
- Ensure `.env` is in your current working directory
- Check file permissions (should be readable)
- Verify the format: `KEY=value` (no spaces around `=`)

### Colors look wrong
- Try a different terminal emulator
- Ensure your terminal supports true color (24-bit)
- Try switching theme: `theme = "light"` in config

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.
