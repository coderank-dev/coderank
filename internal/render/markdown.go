package render

import (
	"regexp"

	"github.com/charmbracelet/glamour"
)

// codeFenceLangRe matches fenced code block opening lines with a language identifier.
var codeFenceLangRe = regexp.MustCompile("(?m)^(```|~~~)\\w+")

// stripCodeFenceLanguages removes language identifiers from fenced code blocks
// (e.g. "```typescript" → "```") because Glamour renders the identifier as
// visible plain text rather than using it for syntax highlighting.
func stripCodeFenceLanguages(md string) string {
	return codeFenceLangRe.ReplaceAllStringFunc(md, func(match string) string {
		if match[0] == '`' {
			return "```"
		}
		return "~~~"
	})
}

// RenderMarkdown renders a markdown string with Glamour, using automatic
// dark/light terminal detection and 100-char word wrap. Returns the styled
// string ready for printing. If rendering fails (e.g., non-TTY), returns error.
func RenderMarkdown(md string) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", err
	}
	return r.Render(stripCodeFenceLanguages(md))
}
