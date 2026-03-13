# Charm Ecosystem Skill

## Library Selection Guide

| Need | Library | Example |
|------|---------|---------|
| Styled terminal output | Lip Gloss | Status tables, headers, colored text |
| Render markdown in terminal | Glamour | `coderank fetch` output |
| Interactive forms/wizards | Huh | `coderank init` wizard |
| Full TUI (persistent UI) | Bubble Tea | `coderank shell` REPL only |
| UI components (tables, spinners) | Bubbles | Inside Bubble Tea programs |
| Structured logging | Charm Log | Debug/verbose output |

## Critical Rules
- **Do NOT use Bubble Tea for simple commands.** Most commands just print output and exit.
  - Use Lip Gloss for styling + fmt for output
  - Only use Bubble Tea for `coderank shell` (the REPL)
- **Do NOT use fmt.Println for user-facing output.** Always use Lip Gloss styles.
- **Do use Glamour for markdown rendering.** It's the killer feature for a docs CLI.

## Lip Gloss Patterns

Define styles in `internal/render/styles.go`:
```go
var Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
```

Use them in commands:
```go
fmt.Println(render.Title.Render("Section Header"))
```

## Glamour Patterns

Render markdown with a dark theme:
```go
r, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
out, _ := r.Render(markdownContent)
fmt.Print(out)
```

## Huh Patterns

Build interactive forms:
```go
huh.NewForm(
  huh.NewGroup(
    huh.NewSelect().
      Title("Choose a library").
      Options("react", "vue", "svelte").
      Value(&selected),
  ),
).Run()
```
