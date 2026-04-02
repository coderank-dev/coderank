package agents

// RootSkillMD returns the content of the root-level CodeRank skill.
// This single file (~200 tokens) teaches any agent how to use CodeRank
// for all 300+ indexed libraries — no per-library enumeration needed.
func RootSkillMD() string {
	return `---
name: coderank
description: >-
  Query up-to-date, condensed documentation for 300+ libraries
  (npm, PyPI, Go, Rust, Java, Swift). Use when the developer imports
  a third-party library, asks about an API, needs usage examples,
  or is debugging library-specific issues.
allowed-tools: Bash(coderank *)
---

# CodeRank — Library Documentation for AI Agents

Query condensed, always-fresh documentation for any indexed library.
Pre-processed for AI context windows: only signatures, usage patterns, gotchas, and migration notes.

## Commands

` + "```" + `bash
# Query a specific API — returns signatures, usage, gotchas
coderank query <library> "<api or question>"
# Examples:
#   coderank query react "useCallback vs useMemo"
#   coderank query express "error handling middleware"
#   coderank query gin "route grouping"

# Get a topic overview — grouped APIs and patterns
coderank topic <library> <topic>
# Examples:
#   coderank topic react hooks
#   coderank topic prisma migrations

# Semantic search across all docs for a library
coderank search <library> "<keyword>"

# Get common pitfalls and sharp edges
coderank gotchas <library> [api]

# List available topics for a library
coderank topics <library>
` + "```" + `

## When to Use

- **Developer imports a library** → ` + "`coderank query`" + ` for the API they're using
- **Working in a feature area** → ` + "`coderank topic`" + ` for grouped context
- **Debugging or troubleshooting** → ` + "`coderank gotchas`" + ` for known pitfalls
- **Not sure where to look** → ` + "`coderank search`" + ` for semantic search
- **Exploring what's available** → ` + "`coderank topics`" + ` to list topic areas

## Notes

- Output is plain markdown on stdout, errors on stderr
- Queries hit pre-indexed docs (sub-second response, no live fetching)
- Works offline if the library was previously cached
- Use ` + "`coderank query <library> --help`" + ` if unsure about a library name
`
}
