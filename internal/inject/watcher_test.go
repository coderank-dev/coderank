package inject

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchedFilesCoversAllEcosystems(t *testing.T) {
	// Verify we're watching the key dependency files for every ecosystem
	expected := map[string]bool{
		"package.json":     false, // npm/JS
		"go.mod":           false, // Go
		"requirements.txt": false, // Python
		"Cargo.toml":       false, // Rust
		".coderank.yml":    false, // CodeRank config
	}

	for _, file := range WatchedFiles {
		if _, ok := expected[file]; ok {
			expected[file] = true
		}
	}

	for file, found := range expected {
		assert.True(t, found,
			"WatchedFiles should include %s for its ecosystem", file)
	}
}

func TestWatchReturnsErrorWhenNoFilesFound(t *testing.T) {
	// Empty directory — no dependency files to watch
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := Watch(ctx, dir, func() error { return nil })
	assert.ErrorContains(t, err, "no dependency files",
		"should error when there are no watchable files")
}

func TestWatchDetectsFileInProjectDir(t *testing.T) {
	dir := t.TempDir()
	// Create a package.json so the watcher has something to monitor
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "package.json"),
		[]byte(`{"dependencies":{}}`), 0644,
	))

	// We can't easily test the full watch loop in a unit test (it blocks),
	// but we can verify the setup doesn't error when files exist.
	// The watch loop itself is tested manually and in integration tests.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so Watch returns

	// Watch should start successfully (find the file) then exit on cancel
	err := Watch(ctx, dir, func() error { return nil })
	assert.NoError(t, err)
}
