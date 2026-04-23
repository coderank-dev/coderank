package agents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteMarkerSectionCreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")

	require.NoError(t, WriteMarkerSection(path, "# CodeRank\nhello\n"))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	str := string(got)
	assert.Contains(t, str, MarkerStart)
	assert.Contains(t, str, MarkerEnd)
	assert.Contains(t, str, "# CodeRank\nhello")
}

func TestWriteMarkerSectionReplacesExistingSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	existing := "# Custom rules above\n\n" +
		MarkerStart + "\n" +
		"old body\n" +
		MarkerEnd + "\n" +
		"\n# Custom rules below\n"
	require.NoError(t, os.WriteFile(path, []byte(existing), 0644))

	require.NoError(t, WriteMarkerSection(path, "new body"))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	str := string(got)
	assert.Contains(t, str, "Custom rules above", "content before start marker must be preserved")
	assert.Contains(t, str, "Custom rules below", "content after end marker must be preserved")
	assert.Contains(t, str, "new body", "new content must be written")
	assert.NotContains(t, str, "old body", "old content between markers must be gone")
}

func TestWriteMarkerSectionAppendsToFileWithoutMarkers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".windsurfrules")
	require.NoError(t, os.WriteFile(path, []byte("# Existing rules\n"), 0644))

	require.NoError(t, WriteMarkerSection(path, "coderank body"))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	str := string(got)
	assert.Contains(t, str, "Existing rules", "pre-existing content must be preserved")
	assert.Contains(t, str, MarkerStart, "start marker appended")
	assert.Contains(t, str, MarkerEnd, "end marker appended")
	assert.Contains(t, str, "coderank body", "coderank content appended")
}
