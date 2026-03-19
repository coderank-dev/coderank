package inject

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteContextCreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	agent := Agent{Name: "Claude Code", ContextPath: ".claude/context/coderank-stack.md"}
	content := ContextContent{
		Libraries:   []string{"react@19.1", "prisma@6.0"},
		Body:        "# React 19.1 — API Surface\n...\n",
		TotalTokens: 1500,
	}

	err := WriteContext(dir, agent, content)
	require.NoError(t, err)

	written, err := os.ReadFile(filepath.Join(dir, agent.ContextPath))
	require.NoError(t, err)
	assert.Contains(t, string(written), "react@19.1, prisma@6.0",
		"header should list all injected libraries")
	assert.Contains(t, string(written), "API Surface",
		"body content should be included")
	assert.Contains(t, string(written), "coderank inject",
		"header should credit coderank inject so users know not to edit manually")
}

func TestWriteContextReplacesMarkerSection(t *testing.T) {
	dir := t.TempDir()

	// Simulate an existing AGENTS.md with other content + old CodeRank section
	agentsPath := filepath.Join(dir, "AGENTS.md")
	existingContent := `# My Agent Rules

Some custom rules here.

<!-- coderank:start -->
<!-- old content -->
# Old API Surface
<!-- coderank:end -->

More custom rules.
`
	require.NoError(t, os.WriteFile(agentsPath, []byte(existingContent), 0644))

	agent := Agent{Name: "Codex", ContextPath: "AGENTS.md"}
	content := ContextContent{
		Libraries:   []string{"hono@4.0"},
		Body:        "# Hono 4.0 — API Surface\n",
		TotalTokens: 800,
	}

	err := WriteContext(dir, agent, content)
	require.NoError(t, err)

	updated, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	str := string(updated)
	assert.Contains(t, str, "Some custom rules here",
		"content before the marker should be preserved")
	assert.Contains(t, str, "Hono 4.0",
		"new API surface should replace the old one")
	assert.NotContains(t, str, "Old API Surface",
		"old CodeRank content should be removed")
	assert.Contains(t, str, "More custom rules",
		"content after the marker should be preserved")
}

func TestWriteContextAppendsToFileWithoutMarkers(t *testing.T) {
	dir := t.TempDir()

	// Existing .windsurfrules with no CodeRank section
	rulesPath := filepath.Join(dir, ".windsurfrules")
	require.NoError(t, os.WriteFile(rulesPath, []byte("# Existing rules\n"), 0644))

	agent := Agent{Name: "Windsurf", ContextPath: ".windsurfrules"}
	content := ContextContent{
		Libraries:   []string{"zod@3.24"},
		Body:        "# Zod 3.24 — API Surface\n",
		TotalTokens: 500,
	}

	err := WriteContext(dir, agent, content)
	require.NoError(t, err)

	updated, err := os.ReadFile(rulesPath)
	require.NoError(t, err)
	str := string(updated)
	assert.Contains(t, str, "Existing rules",
		"original content should be preserved")
	assert.Contains(t, str, "Zod 3.24",
		"CodeRank content should be appended")
}
