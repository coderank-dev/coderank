package wiki

import "strings"

// MatchLibraries returns the subset of `candidates` whose name appears as a
// case-insensitive substring of `text`. Used by the wiki hooks to decide
// whether a user prompt or edited file is library-related enough to warrant
// a coderank reminder. Intentionally permissive: false positives are a mild
// nuisance (an unneeded nudge), missed matches would hurt reliability.
func MatchLibraries(text string, candidates []string) []string {
	lower := strings.ToLower(text)
	var found []string
	seen := map[string]bool{}
	for _, lib := range candidates {
		name := strings.ToLower(strings.TrimSpace(lib))
		if name == "" || seen[name] {
			continue
		}
		if strings.Contains(lower, name) {
			found = append(found, lib)
			seen[name] = true
		}
	}
	return found
}
