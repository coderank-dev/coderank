package agents

import (
	"fmt"
	"os"
	"strings"
)

// Default marker delimiters used by `coderank inject` for the single shared
// section of context files like AGENTS.md or .windsurfrules. The harness
// emitter uses per-skill markers via WriteMarkerSectionCustom so multiple
// coderank sections (root, wiki, ...) coexist without overwriting each other.
const (
	MarkerStart = "<!-- coderank:start -->"
	MarkerEnd   = "<!-- coderank:end -->"
)

// WriteMarkerSection is the back-compat entry point used by `coderank inject`.
// It writes content between the default MarkerStart/MarkerEnd delimiters.
// For multiple distinct sections in the same file, callers should use
// WriteMarkerSectionCustom with per-section markers instead.
func WriteMarkerSection(path, content string) error {
	return WriteMarkerSectionCustom(path, content, MarkerStart, MarkerEnd)
}

// WriteMarkerSectionCustom writes content between the given start/end markers
// in the file at path. Behavior:
//
//   - If the file doesn't exist, it's created with the content wrapped in markers.
//   - If the file exists but lacks this marker pair, the marked section is
//     appended, preserving all existing content (including any other coderank
//     marker sections).
//   - If the file has this exact marker pair, the content between them is
//     replaced in place; everything before the start marker and after the end
//     marker is preserved.
//   - If the file has a start marker but no matching end marker (corruption),
//     a new marked section is appended rather than eating the file.
//
// Multiple distinct marker pairs (e.g. coderank:coderank and coderank:wiki)
// can coexist in the same file: each call only touches its own pair.
func WriteMarkerSectionCustom(path, content, startMarker, endMarker string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	markedContent := startMarker + "\n" + content + "\n" + endMarker + "\n"

	if len(existing) == 0 {
		return os.WriteFile(path, []byte(markedContent), 0644)
	}

	existingStr := string(existing)
	before, rest, found := strings.Cut(existingStr, startMarker)
	if !found {
		return os.WriteFile(path, []byte(existingStr+"\n"+markedContent), 0644)
	}

	_, after, found := strings.Cut(rest, endMarker+"\n")
	if !found {
		_, after, found = strings.Cut(rest, endMarker)
	}
	if !found {
		return os.WriteFile(path, []byte(existingStr+"\n"+markedContent), 0644)
	}

	return os.WriteFile(path, []byte(before+markedContent+after), 0644)
}
