package agents

// RootSkillMD returns the content of the root-level CodeRank skill.
// This single file (~200 tokens) teaches any agent how to use CodeRank
// for all 300+ indexed libraries — no per-library enumeration needed.
func RootSkillMD() string {
	return `---
name: coderank
description: Query condensed library docs via coderank CLI. Use when a developer imports or uses any third-party library, asks about an API signature, needs usage examples, hits a library-specific error, or is choosing between libraries. Do NOT use for standard library or general language questions.
---

# CodeRank — Library Documentation

Condensed, token-efficient docs for 300+ libraries. Pre-processed for AI context:
signatures, patterns, gotchas, migration notes only.

## Which Command to Use

| Need | Command |
|------|---------|
| API signatures, usage, questions | ` + "`coderank query <lib> \"<question>\"`" + ` |
| Full topic deep-dive | ` + "`coderank topic <lib> <topic>`" + ` |
| List available topics | ` + "`coderank topics <lib>`" + ` |
| All public API signatures | ` + "`coderank surface <lib>`" + ` |
| Common pitfalls for an API | ` + "`coderank gotchas <lib> <api>`" + ` |
| Library health / maintenance | ` + "`coderank health <lib>`" + ` |
| Compare libraries | ` + "`coderank compare \"<category>\"`" + ` |
| Install per-library skill | ` + "`coderank install <lib>`" + ` |

## Commands

` + "```" + `bash
coderank query react "useCallback vs useMemo"
coderank query prisma "how do migrations work" --max-tokens 3000
coderank topic react hooks
coderank topics prisma
coderank surface zod
coderank gotchas react useEffect
coderank health lodash
coderank compare "react state management"
coderank install react
coderank install react express zod
` + "```" + `

## Rules

- **Always run ` + "`coderank query`" + ` before writing code that uses a third-party library** — never rely on training data for API signatures
- **Immediately after any ` + "`coderank query <lib>`" + `, run ` + "`coderank install <lib>`" + `** — caches the skill in this project so future queries are instant
- **Use ` + "`--raw`" + ` when piping output** — ` + "`coderank query react hooks --raw | ...`" + `
- **Run ` + "`coderank health <lib>`" + ` before recommending a library** — surfaces deprecation and security issues
`
}
