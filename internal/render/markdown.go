package render

import "github.com/charmbracelet/glamour"

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
	return r.Render(md)
}
