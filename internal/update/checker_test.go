package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureV(t *testing.T) {
	assert.Equal(t, "v1.2.3", ensureV("1.2.3"), "should add v prefix when missing")
	assert.Equal(t, "v1.2.3", ensureV("v1.2.3"), "should not double-add v prefix")
}

func TestCheckSkipsDevBuild(t *testing.T) {
	result := Check("dev")
	assert.Nil(t, result, "dev builds should skip version check")

	result = Check("")
	assert.Nil(t, result, "empty version should skip version check")
}

func TestNoticeStringUpdateAvailable(t *testing.T) {
	result := &CheckResult{
		CurrentVersion: "0.1.0",
		LatestVersion:  "0.2.0",
		UpdateAvail:    true,
	}
	notice := result.NoticeString()
	assert.Contains(t, notice, "0.2.0 is available", "notice should include new version")
	assert.Contains(t, notice, "you have 0.1.0", "notice should include current version")
	assert.Contains(t, notice, "coderank update", "notice should include update command")
}

func TestNoticeStringNoUpdate(t *testing.T) {
	var result *CheckResult
	assert.Equal(t, "", result.NoticeString(), "nil result should return empty string")

	result = &CheckResult{UpdateAvail: false}
	assert.Equal(t, "", result.NoticeString(), "no update available should return empty string")
}

func TestCacheRoundTrip(t *testing.T) {
	// Override cachePath so this test doesn't touch the real config dir.
	dir := t.TempDir()
	orig := cachePath
	cachePath = func() string {
		return filepath.Join(dir, "coderank", "update-check.json")
	}
	defer func() { cachePath = orig }()

	// Nothing cached yet.
	_, ok := loadCache()
	assert.False(t, ok, "empty cache dir should return no result")

	// Save and reload.
	saveCache(cachedCheck{
		CheckedAt:     time.Now(),
		LatestVersion: "0.3.0",
		ReleaseURL:    "https://github.com/coderank-dev/coderank/releases/tag/v0.3.0",
	})

	data, err := os.ReadFile(cachePath())
	require.NoError(t, err, "cache file should exist after saveCache")

	var loaded cachedCheck
	require.NoError(t, json.Unmarshal(data, &loaded))
	assert.Equal(t, "0.3.0", loaded.LatestVersion, "loaded version should match saved version")

	cached, ok := loadCache()
	require.True(t, ok, "fresh cache should be valid")
	assert.Equal(t, "0.3.0", cached.LatestVersion)
}

func TestIsHomebrewDetection(t *testing.T) {
	cases := []struct {
		path     string
		expected bool
	}{
		{"/usr/local/Cellar/coderank/0.1.0/bin/coderank", true},
		{"/opt/homebrew/Cellar/coderank/0.1.0/bin/coderank", true},
		{"/home/user/.local/bin/coderank", false},
		{"/usr/local/bin/coderank", false},
	}

	for _, tc := range cases {
		result := strings.Contains(tc.path, "/Cellar/") || strings.Contains(tc.path, "/homebrew/")
		assert.Equal(t, tc.expected, result, "path: %s", tc.path)
	}
}
