package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripCodeFenceLanguages(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes typescript identifier",
			input: "```typescript\nconst x = 1\n```",
			want:  "```\nconst x = 1\n```",
		},
		{
			name:  "removes go identifier",
			input: "```go\nfmt.Println()\n```",
			want:  "```\nfmt.Println()\n```",
		},
		{
			name:  "preserves plain fences",
			input: "```\nplain\n```",
			want:  "```\nplain\n```",
		},
		{
			name:  "handles tilde fences",
			input: "~~~bash\necho hi\n~~~",
			want:  "~~~\necho hi\n~~~",
		},
		{
			name:  "strips language and indent from 4-space indented fence",
			input: "    ```typescript\n    const x = 1\n    ```",
			want:  "```\n    const x = 1\n```",
		},
		{
			name:  "de-indents plain 4-space fence closing",
			input: "```typescript\nconst x = 1\n    ```",
			want:  "```\nconst x = 1\n```",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, stripCodeFenceLanguages(tc.input))
		})
	}
}

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
