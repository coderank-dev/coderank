// Package agentdetect provides shared detection logic for AI coding agents.
// Both the inject (context injection) and agents (skill installation) packages
// use this to discover which agents are present in a project directory.
package agentdetect

import (
	"os"
	"path/filepath"
)

// Indicator describes how to detect a single agent's presence.
type Indicator struct {
	// Path is the file or directory to check (relative to project root).
	Path string
	// IsDir indicates whether Path should be a directory (true) or file (false).
	IsDir bool
}

// IsPresent reports whether this indicator's path exists in projectRoot.
func (ind Indicator) IsPresent(projectRoot string) bool {
	full := filepath.Join(projectRoot, ind.Path)
	info, err := os.Stat(full)
	if err != nil {
		return false
	}
	if ind.IsDir {
		return info.IsDir()
	}
	return !info.IsDir()
}
