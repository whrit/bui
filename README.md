# bui

A terminal-native GitHub Pull Request manager (create/review/manage/merge) with **built-in LLM copilot**, **syntax-highlighted diffs**, and a **mandatory TOML config**.

This repo is a **starter scaffold**:
- Config system (auto-generates `config.toml` on first run)
- LLM provider abstraction + OpenAI/Anthropic/Local providers (stubs + real HTTP wiring)
- Syntax highlighting pipeline for diffs (Chroma-based)
- Bubble Tea app skeleton (screen router + placeholder dashboard)

## Requirements
- Go 1.22+
- GitHub CLI (`gh`) installed and authenticated

## Quick start
```bash
go mod download
go run ./cmd/bui
```

Config will be generated at:
- macOS/Linux: `~/.config/bui/config.toml`
- Windows: `%APPDATA%\bui\config.toml`

## LLM setup
Set one of:
- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- or use local provider (see config)

Then configure `[llm]` in `config.toml`.

## Notes
- This starter focuses on *architecture + scaffolding*. The PR lifecycle screens are intentionally thin; build features iteratively.
