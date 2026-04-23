package agents

// WikiSkillMD returns the content of the per-project CodeRank wiki skill.
// Installed by 'coderank init' into detected agents. Instructs the agent
// to maintain and consult a project-local knowledge base at .coderank/wiki/
// using the `coderank wiki` CLI commands. Each workflow step is a single
// deterministic tool call rather than a multi-step markdown-editing ritual.
func WikiSkillMD() string {
	return `---
name: coderank-wiki
description: Consult and maintain the project-specific knowledge wiki at .coderank/wiki/. ALWAYS invoke in two situations: (1) BEFORE answering any question about a third-party library used in this project - wiki pages capture project-specific patterns, decisions, and gotchas that override generic docs; (2) AFTER writing or modifying code that adopts a new library, introduces a new pattern, resolves a library-specific bug, or makes an architectural decision - record what was done so the knowledge persists. Do not skip the ingest step when you have changed code that uses a library. Wiki content reflects actual project decisions and supersedes training data.
---

# Project Wiki

A short, project-specific knowledge base at ` + "`.coderank/wiki/`" + `. Managed via the
` + "`coderank wiki`" + ` CLI, not by editing markdown directly.

## What belongs here (short, pointer-style entries)

Wiki pages are NOT library documentation. They record how THIS project uses
a library, with file pointers. Library docs are served by ` + "`coderank query <lib>`" + `;
don't duplicate them here.

A good page is 3-8 lines of body plus a ` + "`--refs`" + ` list pointing at the relevant
source files:

- one-sentence summary of the pattern or decision
- list of key files (via ` + "`--refs`" + `, surfaced automatically in the page frontmatter)
- any non-obvious gotcha, rename history, or constraint

If you find yourself writing paragraphs of library explanation, stop -
that's ` + "`coderank query`" + ` territory.

## Query (before answering library questions)

Run these before answering any library question about this project:

` + "```" + `bash
coderank wiki list                   # see what pages exist
coderank wiki read <lib> <topic>     # read a matching page
` + "```" + `

Wiki pages reflect actual project decisions and may override generic library
docs. If both apply, combine them in your answer.

## Ingest (after changing code that uses a library)

Run after: adopting a new library, introducing a new pattern, resolving a
library-specific bug, or making an architectural decision that future agents
should know about. Single atomic command:

` + "```" + `bash
coderank wiki ingest \
  --lib <library> \
  --topic <short-topic-slug> \
  --summary "one-to-three sentence summary" \
  --refs "src/path/one.ts,src/path/two.ts" \
  --description "one-line description for the index"
` + "```" + `

For longer bodies, pipe via ` + "`--body-from-stdin`" + ` or read from a file with
` + "`--body-from-file <path>`" + `. One invocation writes the page, updates index.md,
and appends an [INGEST] entry to log.md.

Skip ingest only when the change does NOT represent new project knowledge -
simple bug fixes in your own code, formatting, or refactors that don't change
how a library is used.

## Lint (periodic cleanup)

When ` + "`log.md`" + ` has accumulated many [INGEST] entries without a recent [LINT],
or when you suspect drift (pages referencing removed files, deprecated APIs),
run:

` + "```" + `bash
coderank wiki lint
` + "```" + `

Lint scans for missing pages (listed but not on disk), orphans (on disk but
not listed), and pages marked ` + "`status: deprecated`" + `. Returns non-zero when
issues are found.

## Page format (handled by the CLI - do not edit by hand)

The CLI writes a frontmatter-prefixed markdown file like:

` + "```" + `markdown
---
status: current
updated: 2026-04-23
related: [react/hooks]
refs: [src/schemas/auth.ts, src/api/auth.ts]
description: Auth schemas with discriminated unions
---
# zod - auth-schemas

[body]
` + "```" + `

You do not need to author this frontmatter yourself; ` + "`coderank wiki ingest`" + `
renders it from flags. If you need to edit an existing page, re-ingest with
the same ` + "`--lib`" + ` / ` + "`--topic`" + ` - the page is replaced, the index is updated in
place, and the log gets a new entry.

## Rules

- Always query the wiki (` + "`coderank wiki list`" + `) before answering library questions
- Always ingest after changing code that uses a library in a non-trivial way
- Keep pages short and pointer-style; link to code via ` + "`--refs`" + `
- Never edit ` + "`.coderank/wiki/index.md`" + ` or ` + "`log.md`" + ` by hand - the CLI manages them
- One library + one topic per page
`
}
