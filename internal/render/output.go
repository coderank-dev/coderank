package render

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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
// and token count. Displayed above the Glamour-rendered markdown content.
func DocHeader(library, version, topic string, tokens int) string {
	var parts []string
	parts = append(parts, Title.Render(library+"@"+version))
	if topic != "" && topic != "_api-surface" {
		parts = append(parts, Subtle.Render(" → "+topic))
	}
	if tokens > 0 {
		parts = append(parts, Subtle.Render(fmt.Sprintf(" (%d tokens)", tokens)))
	}
	return strings.Join(parts, "") + "\n"
}

// DocFooter renders a styled footer with query metadata.
func DocFooter(totalTokens, queryMs int) string {
	return Subtle.Render(fmt.Sprintf(
		"\n─── %d tokens · %dms ───",
		totalTokens, queryMs,
	)) + "\n"
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
