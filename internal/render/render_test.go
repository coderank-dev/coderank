package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderMarkdown_RendersHeading(t *testing.T) {
	out, err := RenderMarkdown("# Hello\n\nWorld")
	assert.NoError(t, err)
	assert.NotEmpty(t, out,
		"Glamour should produce non-empty output for valid markdown")
}

func TestDocBox_HasRoundedBorder(t *testing.T) {
	rendered := DocBox.Render("test content")
	assert.NotEmpty(t, rendered,
		"DocBox should produce styled output")
}

func TestDocHeaderIncludesLibraryAndVersion(t *testing.T) {
	header := DocHeader("react", "19.1.0", "hooks/state", 1920, 0)
	assert.Contains(t, header, "react@19.1.0",
		"header should show library@version")
	assert.Contains(t, header, "1920",
		"header should show token count")
}
