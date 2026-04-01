package render

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
)

// codeFenceLangRe matches any fenced code block fence line, including lines with
// leading whitespace (e.g. LLM-generated "    ```typescript" from the "indented 4
// spaces" template instruction). Captures: leading whitespace, fence chars, language.
var codeFenceLangRe = regexp.MustCompile("(?m)^[ \t]*(`{3,}|~{3,})\\w*[ \t]*$")

// stripCodeFenceLanguages normalises fenced code block fence lines:
//   - removes language identifiers (Glamour shows them as visible plain text)
//   - strips leading whitespace from fence lines (4-space-indented fences are treated
//     by Markdown parsers as indented code blocks, making the backticks render literally)
func stripCodeFenceLanguages(md string) string {
	return codeFenceLangRe.ReplaceAllStringFunc(md, func(match string) string {
		if strings.Contains(match, "`") {
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
