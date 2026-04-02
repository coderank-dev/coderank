package render

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StripFrontmatter removes the YAML frontmatter block (--- ... ---) from
// markdown content before rendering. Returns the body unchanged if no
// frontmatter is found.
func StripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return content
	}
	rest := content[4:]
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return content
	}
	return strings.TrimLeft(rest[end+5:], "\n")
}

// Markdown outputs plain markdown to stdout (for piping to agents).
// This is the preferred output for agent consumption — more compact than JSON.
func Markdown(md string) {
	fmt.Print(md)
}

// JSON outputs structured data as JSON to stdout.
func JSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// Rendered outputs Glamour-rendered markdown with Lip Gloss chrome to stdout.
// Falls back to raw markdown if Glamour fails (CI, non-TTY).
func Rendered(md string, title string, statusLine string) {
	rendered, err := RenderMarkdown(md)
	if err != nil {
		// Glamour can fail in non-TTY environments.
		// Fall back to raw markdown — still useful, just not pretty.
		Markdown(md)
		return
	}

	box := DocBox.Render(rendered)
	if title != "" {
		fmt.Println(Title.Render(title))
	}
	fmt.Println(box)
	if statusLine != "" {
		fmt.Println(StatusLine.Render(statusLine))
	}
}

// DocHeader renders a styled header showing library name, version, topic,
// token count, and relevance score. Displayed above the Glamour-rendered markdown content.
// score is 0–100; pass 0 to omit it.
func DocHeader(library, version, topic string, tokens, score int) string {
	var parts []string

	lib := "📦 " + Title.Render(library)
	if version != "" {
		lib += Subtle.Render(" "+version)
	}
	parts = append(parts, lib)

	if topic != "" && topic != "_api-surface" {
		parts = append(parts, Subtle.Render("  →  "+topic))
	}

	var meta []string
	if tokens > 0 {
		meta = append(meta, fmt.Sprintf("%s tokens", formatInt(tokens)))
	}
	if score > 0 {
		meta = append(meta, fmt.Sprintf("%d%% match", score))
	}
	if len(meta) > 0 {
		parts = append(parts, Subtle.Render("  ·  "+strings.Join(meta, "  ·  ")))
	}

	return strings.Join(parts, "") + "\n"
}

// DocFooter renders a styled footer with query metadata.
func DocFooter(totalTokens, queryMs int) string {
	timing := lipgloss.NewStyle().Foreground(Accent).Bold(true).Render(fmt.Sprintf("⚡ %dms", queryMs))
	if totalTokens <= 0 {
		return "\n" + timing + "\n"
	}
	tokens := Subtle.Render(fmt.Sprintf("%s tokens total", formatInt(totalTokens)))
	return "\n" + timing + Subtle.Render("  ·  ") + tokens + "\n"
}

// formatInt formats an integer with comma separators (e.g. 2847 → "2,847").
func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

// ErrorMsg renders a styled error message to stderr. Red and bold.
// Use this instead of fmt.Fprintf(os.Stderr, ...).
func ErrorMsg(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, Error.Render("✗ "+msg))
}

// SuccessMsg renders a styled success message. Green.
func SuccessMsg(msg string) string {
	return Success.Render("✓ " + msg) + "\n"
}

// WarningMsg renders a styled warning. Yellow.
func WarningMsg(msg string) string {
	return Warning.Render("⚠ " + msg) + "\n"
}

// InfoMsg outputs a styled info message to stderr (for non-data output).
func InfoMsg(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, MutedText.Render(msg))
}
