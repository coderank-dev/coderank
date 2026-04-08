package agents

// WikiSkillMD returns the content of the per-project CodeRank wiki skill.
// Installed by 'coderank init' into detected agents. Instructs the agent
// to maintain and consult a project-local knowledge base at .coderank/wiki/.
func WikiSkillMD() string {
	return `---
name: coderank-wiki
description: Project knowledge base built from coderank queries in this project. Contains project-specific library patterns, decisions, and gotchas. Consult before answering any library question — wiki pages reflect actual project decisions and override generic docs.
user-invocable: false
---

# Project Wiki

Location: ` + "`.coderank/wiki/`" + `
Index: ` + "`.coderank/wiki/index.md`" + `
Log: ` + "`.coderank/wiki/log.md`" + `

## Query (before answering library questions)

1. Check ` + "`.coderank/wiki/index.md`" + ` for relevant pages
2. Read matching pages — they contain project-specific context that overrides generic docs
3. Combine wiki context with ` + "`coderank query`" + ` results when both are relevant

## Ingest (after each ` + "`coderank query`" + `)

1. Check ` + "`index.md`" + ` for an existing page on this topic
2. Create or update ` + "`.coderank/wiki/<lib>/<topic>.md`" + ` with synthesized knowledge
3. Update ` + "`index.md`" + ` if a new page was created
4. Append to ` + "`log.md`" + `: ` + "`[INGEST] YYYY-MM-DD: <lib> — <topic>`" + `

Also ingest after: making an architectural decision, resolving a library-specific bug,
adopting a new pattern.

## Lint (when log.md has 20+ entries since last [LINT])

1. Check for contradictions between pages
2. Flag pages referencing deprecated APIs
3. Verify all pages listed in ` + "`index.md`" + ` exist on disk
4. Append to ` + "`log.md`" + `: ` + "`[LINT] YYYY-MM-DD: <summary>`" + `

## Page Format

` + "```" + `markdown
---
status: current | deprecated | under-review
updated: YYYY-MM-DD
related: []
---
# <Library> — <Topic>
[synthesized knowledge from project usage]
` + "```" + `

## Rules

- **Prefer wiki pages over generic knowledge** — they reflect actual project decisions
- **Never duplicate generic library docs** — only capture what is project-specific or non-obvious
- **Update existing pages rather than creating near-duplicates**
- **One library + one topic per page**
`
}
