package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestCache(t *testing.T) *Manager {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mgr, err := NewManager()
	require.NoError(t, err)
	t.Cleanup(func() { mgr.Close() })
	return mgr
}

func TestPutAndSearch(t *testing.T) {
	mgr := setupTestCache(t)

	err := mgr.Put("react", "19.1.0", "hooks-state", 1920, []byte("# React Hooks"))
	require.NoError(t, err)

	results, err := mgr.Search("react hooks", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "react", results[0].Library)
	assert.Equal(t, "hooks-state", results[0].Topic)
}

func TestPutWritesFileToDisk(t *testing.T) {
	mgr := setupTestCache(t)

	content := []byte("# React Hooks State")
	err := mgr.Put("react", "19.1.0", "hooks-state", 1920, content)
	require.NoError(t, err)

	home, _ := os.UserHomeDir()
	filePath := filepath.Join(home, ".coderank", "cache", "react", "19.1.0", "hooks-state.md")
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestStats(t *testing.T) {
	mgr := setupTestCache(t)

	mgr.Put("react", "19.1.0", "hooks", 1920, []byte("hooks"))
	mgr.Put("react", "19.1.0", "components", 2100, []byte("components"))

	count, tokens, err := mgr.Stats()
	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, 4020, tokens)
}
