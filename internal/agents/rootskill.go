package agents

// RootSkillMD returns the content of the root-level CodeRank skill.
// This single file (~200 tokens) teaches any agent how to use CodeRank
// for all 300+ indexed libraries — no per-library enumeration needed.
func RootSkillMD() string {
	return `---
name: coderank
description: "Query condensed docs for 300+ libraries via coderank CLI. TRIGGER when: developer imports a third-party library, asks about an API, needs usage examples, or is debugging library-specific issues. DO NOT TRIGGER for standard library or general programming questions."
---

# CodeRank — Library Documentation for AI Agents

Query condensed, always-fresh documentation for any indexed library.
Pre-processed for AI context windows: only signatures, usage patterns, gotchas, and migration notes.

## Documentation Commands

` + "```" + `bash
# Query a specific API — returns signatures, usage, gotchas
coderank query <library> "<api or question>"
# Examples:
#   coderank query react "useCallback vs useMemo"
#   coderank query express "error handling middleware"
#   coderank query gin "route grouping"

# Get full topic content by name
coderank query <library> <topic>
# Examples:
#   coderank query react hooks
#   coderank query prisma migrations

# Show compact API surface (all public signatures)
coderank surface <library>

# Get health score and maintenance status
coderank health <library>

# Compare libraries in a category
coderank compare <category>
# Examples:
#   coderank compare "react state management"
` + "```" + `

## Agent Integration Commands

` + "```" + `bash
# Install CodeRank skill into detected AI coding agents (run once per project)
coderank install
coderank install --global                        # Install to all projects
coderank install --with-surfaces react,express   # Also add per-library API surfaces
coderank install --agents claude,cursor          # Target specific agents
coderank install --dry-run                       # Preview without writing files

# Pre-inject API surfaces into agent context (ambient docs before agent starts)
coderank inject react express typescript
coderank inject --target claude                  # Target specific agent
coderank inject --global                         # Write to global context
` + "```" + `

## Setup Commands

` + "```" + `bash
# Authenticate with API key
coderank auth <api-key>

# Log in via browser
coderank login

# Manage offline cache
coderank cache sync <library>    # Download for offline use
coderank cache list              # Show cached libraries
coderank cache clear <library>   # Remove from cache
` + "```" + `

## When to Use

- **Developer imports a library** → ` + "`coderank query`" + ` for the API they're using
- **Working in a feature area** → ` + "`coderank query`" + ` for grouped context
- **Need all public APIs at a glance** → ` + "`coderank surface`" + ` for compact signatures
- **Checking library quality** → ` + "`coderank health`" + ` for maintenance score
- **Choosing between libraries** → ` + "`coderank compare`" + ` for ranked alternatives
- **Setting up a new project** → ` + "`coderank inject`" + ` to pre-load stack docs into agent context

## Notes

- Output is plain markdown on stdout, errors on stderr
- Queries hit pre-indexed docs (sub-second response, no live fetching)
- Works offline for cached libraries (` + "`coderank cache sync`" + ` first)
- Run ` + "`coderank install`" + ` once per project to keep skill docs current
`
}
