package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, err := Load(path)
	require.NoError(t, err)
	assert.Empty(t, s.Hooks)
	assert.Empty(t, s.Passthrough)
}

func TestAddAndSaveCreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, err := Load(path)
	require.NoError(t, err)
	s.AddCoderankHook(EventUserPromptSubmit, "", "coderank wiki hook user-prompt")
	s.AddCoderankHook(EventPostToolUse, "Edit|Write|MultiEdit", "coderank wiki hook post-edit")
	require.NoError(t, s.Save(path))

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, json.Unmarshal(b, &got))
	hooks := got["hooks"].(map[string]any)
	assert.Contains(t, hooks, EventUserPromptSubmit)
	assert.Contains(t, hooks, EventPostToolUse)
}

func TestAddIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, _ := Load(path)
	s.AddCoderankHook(EventPostToolUse, "Edit|Write", "coderank wiki hook post-edit")
	s.AddCoderankHook(EventPostToolUse, "Edit|Write", "coderank wiki hook post-edit")
	require.NoError(t, s.Save(path))

	s2, err := Load(path)
	require.NoError(t, err)
	assert.Len(t, s2.Hooks[EventPostToolUse], 1, "same matcher+command must not duplicate")
	assert.Len(t, s2.Hooks[EventPostToolUse][0].Hooks, 1)
}

func TestAddUpdatesExistingCoderankEntryInPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, _ := Load(path)
	s.AddCoderankHook(EventPostToolUse, "Edit|Write", "coderank wiki hook post-edit")
	require.NoError(t, s.Save(path))

	s2, _ := Load(path)
	s2.AddCoderankHook(EventPostToolUse, "Edit|Write", "coderank wiki hook post-edit --v2")
	require.NoError(t, s2.Save(path))

	s3, _ := Load(path)
	require.Len(t, s3.Hooks[EventPostToolUse], 1)
	require.Len(t, s3.Hooks[EventPostToolUse][0].Hooks, 1)
	assert.Equal(t, "coderank wiki hook post-edit --v2", s3.Hooks[EventPostToolUse][0].Hooks[0].Command)
}

func TestAddAppendsToMatcherGroupWithNonCoderankHook(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	existing := `{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {"type": "command", "command": "/path/to/user/lint.sh"}
        ]
      }
    ]
  }
}`
	require.NoError(t, os.WriteFile(path, []byte(existing), 0644))

	s, err := Load(path)
	require.NoError(t, err)
	s.AddCoderankHook(EventPostToolUse, "Edit|Write", "coderank wiki hook post-edit")
	require.NoError(t, s.Save(path))

	s2, _ := Load(path)
	require.Len(t, s2.Hooks[EventPostToolUse], 1)
	require.Len(t, s2.Hooks[EventPostToolUse][0].Hooks, 2, "user's lint hook must be preserved")
	commands := []string{
		s2.Hooks[EventPostToolUse][0].Hooks[0].Command,
		s2.Hooks[EventPostToolUse][0].Hooks[1].Command,
	}
	assert.Contains(t, commands, "/path/to/user/lint.sh")
	assert.Contains(t, commands, "coderank wiki hook post-edit")
}

func TestPassthroughPreservesUnrelatedKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	existing := `{
  "statusLine": {"type": "static", "value": "my status"},
  "hooks": {}
}`
	require.NoError(t, os.WriteFile(path, []byte(existing), 0644))

	s, err := Load(path)
	require.NoError(t, err)
	s.AddCoderankHook(EventPostToolUse, "Edit", "coderank wiki hook post-edit")
	require.NoError(t, s.Save(path))

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Contains(t, out, "statusLine", "unrelated top-level keys must survive")
}

func TestRemoveOnlyDeletesCoderankEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	existing := `{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {"type": "command", "command": "/path/user/lint.sh"},
          {"type": "command", "command": "coderank wiki hook post-edit"}
        ]
      }
    ]
  }
}`
	require.NoError(t, os.WriteFile(path, []byte(existing), 0644))

	s, err := Load(path)
	require.NoError(t, err)
	s.RemoveCoderankHooks()
	require.NoError(t, s.Save(path))

	s2, _ := Load(path)
	require.Len(t, s2.Hooks[EventPostToolUse], 1)
	require.Len(t, s2.Hooks[EventPostToolUse][0].Hooks, 1)
	assert.Equal(t, "/path/user/lint.sh", s2.Hooks[EventPostToolUse][0].Hooks[0].Command)
}
