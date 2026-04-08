package agents

import "fmt"

// LibrarySkillMD returns the SKILL.md content for a specific library.
// Used by the indexing pipeline to generate per-library skills served
// via GET /v1/skills/:library, and consumed by 'coderank install <lib>'.
//
// Parameters:
//   - lib: library name, e.g. "react"
//   - version: library version, e.g. "18.3"
//   - surface: condensed quick-reference content (top APIs, patterns, gotchas)
//   - paths: glob patterns for files that use this library, e.g. "**/*.ts,**/*.tsx"
func LibrarySkillMD(lib, version, surface, paths string) string {
	return fmt.Sprintf(`---
name: %s
description: %s v%s API — signatures, patterns, gotchas. Use when writing or reviewing code that imports %s.
user-invocable: false
paths: %s
---

# %s v%s

%s

## Need More?

`+"```"+`bash
coderank query %s "<question>"   # semantic search
coderank surface %s              # full API surface
coderank topic %s <topic>        # deep-dive by topic
coderank topics %s               # list available topics
`+"```"+`
`, lib, lib, version, lib, paths, lib, version, surface, lib, lib, lib, lib)
}
