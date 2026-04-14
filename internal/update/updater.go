package update

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// IsHomebrew returns true if the binary appears to be managed by Homebrew.
func IsHomebrew() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.Contains(exe, "/Cellar/") || strings.Contains(exe, "/homebrew/")
}

// SelfUpdate downloads and replaces the current binary with the latest release.
func SelfUpdate() error {
	if IsHomebrew() {
		return fmt.Errorf("installed via Homebrew — run: brew upgrade coderank-dev/tap/coderank")
	}

	latest, _, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	// Asset name matches GoReleaser's default: coderank_{version}_{os}_{arch}.tar.gz
	assetName := fmt.Sprintf("coderank_%s_%s_%s.tar.gz", latest, runtime.GOOS, runtime.GOARCH)
	assetURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s", owner, repo, latest, assetName)

	fmt.Fprintf(os.Stderr, "Downloading CodeRank %s for %s/%s...\n", latest, runtime.GOOS, runtime.GOARCH)

	resp, err := http.Get(assetURL) //nolint:gosec // URL is constructed from known-safe components
	if err != nil {
		return fmt.Errorf("downloading release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("release asset not found: %s (your OS/arch may not be supported)", assetName)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	binaryData, err := extractBinaryFromTarGz(resp.Body, "coderank")
	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	// Resolve the real path of the current executable (follow symlinks).
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating current binary: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	// Atomic replacement: write to a temp file then rename over the original.
	// rename is atomic on POSIX systems; avoids a window where the binary is missing.
	tmpPath := exe + ".tmp"
	if err := os.WriteFile(tmpPath, binaryData, 0755); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s — try: sudo coderank update", exe)
		}
		return fmt.Errorf("writing temp binary: %w", err)
	}

	if err := os.Rename(tmpPath, exe); err != nil {
		os.Remove(tmpPath) //nolint:errcheck // cleanup on failure
		return fmt.Errorf("replacing binary: %w", err)
	}

	// Clear the cache so the update notice disappears immediately on next run.
	os.Remove(cachePath()) //nolint:errcheck // best-effort

	fmt.Fprintf(os.Stderr, "✓ Updated to CodeRank %s\n", latest)
	return nil
}

// extractBinaryFromTarGz reads a .tar.gz archive and returns the contents of
// the first entry whose base name matches name.
func extractBinaryFromTarGz(r io.Reader, name string) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("binary %q not found in archive", name)
		}
		if err != nil {
			return nil, err
		}

		// The binary may live inside a directory within the archive, e.g.
		// coderank_0.2.0_linux_amd64/coderank — match by base name only.
		if filepath.Base(header.Name) == name && header.Typeflag == tar.TypeReg {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}
}
