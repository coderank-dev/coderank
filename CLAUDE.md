# CLAUDE.md — coderank CLI

## What This Repo Does
CodeRank CLI — the user-facing command-line tool for querying AI-optimized library documentation. Built with Go, Cobra (commands), and Charm (terminal UI).

## Tech Stack
- Language: Go 1.22+
- CLI framework: Cobra + Viper (command structure, flags, config cascade)
- Terminal UI: Lip Gloss (styling), Glamour (markdown rendering), Huh (interactive forms)
- Full TUI: Bubble Tea (only for `coderank shell` REPL)
- HTTP client: net/http (API calls to coderank.ai)
- Config: Viper (flags > env > .coderank.yml > defaults)

## Directory Layout
- `cmd/` — One file per Cobra command
- `internal/` — Shared packages (api client, config, render styles)
- `main.go` — Entry point

## Commands
- Build: `go build -o coderank .`
- Test: `go test ./... -v -race`
- Lint: `golangci-lint run`
- Install locally: `go install .`

## Conventions
- Use Lip Gloss for all styled terminal output — never raw fmt.Println for user-facing text
- Use Glamour for rendering markdown content (the core feature for docs display)
- Use Huh forms for interactive wizards (e.g. `coderank init`)
- Wrap all errors with context: `fmt.Errorf("fetch: %w", err)`
- Every command supports `--raw` (plain text), `--json` (JSON), and styled (default) output modes

## Do NOT
- Do NOT use Bubble Tea for simple commands — most commands print output and exit
- Do NOT use global state — pass dependencies through function parameters
- Do NOT use CGO — the binary must be statically linked for Homebrew distribution
- Do NOT call the API without checking for a valid config first
