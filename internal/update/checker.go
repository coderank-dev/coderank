package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	owner         = "coderank-dev"
	repo          = "coderank"
	checkInterval = 24 * time.Hour
	checkTimeout  = 2 * time.Second
)

// CheckResult holds the outcome of a version check.
type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	UpdateAvail    bool
	ReleaseURL     string
}

// cachedCheck is the on-disk format for the update-check.json cache file.
type cachedCheck struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	ReleaseURL    string    `json:"release_url"`
}

// cachePath is a var so tests can override it.
var cachePath = func() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "coderank", "update-check.json")
}

// Check compares the running version against the latest GitHub Release.
// Returns nil if the versions match, the check fails, or a dev build is detected.
func Check(currentVersion string) *CheckResult {
	if currentVersion == "" || currentVersion == "dev" {
		return nil // dev build, skip check
	}

	current := ensureV(currentVersion)

	if cached, ok := loadCache(); ok {
		latest := ensureV(cached.LatestVersion)
		if semver.Compare(current, latest) < 0 {
			return &CheckResult{
				CurrentVersion: currentVersion,
				LatestVersion:  cached.LatestVersion,
				UpdateAvail:    true,
				ReleaseURL:     cached.ReleaseURL,
			}
		}
		return nil // up to date per cache
	}

	latest, releaseURL, err := fetchLatestRelease()
	if err != nil {
		return nil // silently ignore network errors
	}

	saveCache(cachedCheck{
		CheckedAt:     time.Now(),
		LatestVersion: latest,
		ReleaseURL:    releaseURL,
	})

	if semver.Compare(current, ensureV(latest)) < 0 {
		return &CheckResult{
			CurrentVersion: currentVersion,
			LatestVersion:  latest,
			UpdateAvail:    true,
			ReleaseURL:     releaseURL,
		}
	}

	return nil
}

// NoticeString returns a one-line update notice for stderr, or empty string.
func (r *CheckResult) NoticeString() string {
	if r == nil || !r.UpdateAvail {
		return ""
	}
	return fmt.Sprintf("\n⬆ CodeRank %s is available (you have %s). Run: coderank update\n", r.LatestVersion, r.CurrentVersion)
}

// githubRelease is the minimal GitHub Releases API response shape.
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func fetchLatestRelease() (version, releaseURL string, err error) {
	client := &http.Client{Timeout: checkTimeout}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := client.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), release.HTMLURL, nil
}

func loadCache() (cachedCheck, bool) {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return cachedCheck{}, false
	}

	var cached cachedCheck
	if err := json.Unmarshal(data, &cached); err != nil {
		return cachedCheck{}, false
	}

	if time.Since(cached.CheckedAt) > checkInterval {
		return cachedCheck{}, false // stale
	}

	return cached, true
}

func saveCache(c cachedCheck) {
	path := cachePath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.Marshal(c)
	os.WriteFile(path, data, 0644) //nolint:errcheck // best-effort
}

func ensureV(version string) string {
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}
