package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleSkill = `---
name: coderank
description: Query current library documentation via coderank CLI. ALWAYS invoke.
---

# CodeRank - Library Documentation

Body content here.

## Trigger

ALWAYS run ` + "`coderank query`" + ` before answering.
`

func TestEmitSkillSkillMD(t *testing.T) {
	dir := t.TempDir()
	claude, _ := FindByIDs([]string{"claude"})
	require.NoError(t, EmitSkill(dir, claude[0], ScopeProject, "coderank", sampleSkill))

	got, err := os.ReadFile(filepath.Join(dir, ".claude/skills/coderank/SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, sampleSkill, string(got),
		"SkillMD emitter should write the content verbatim (frontmatter kept)")
}

func TestEmitSkillSingleMarkerFile(t *testing.T) {
	dir := t.TempDir()
	codex, _ := FindByIDs([]string{"codex"})
	require.NoError(t, EmitSkill(dir, codex[0], ScopeProject, "coderank", sampleSkill))

	got, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	str := string(got)
	assert.Contains(t, str, "<!-- coderank:coderank:start -->",
		"per-skill start marker must be present so multiple coderank sections can coexist")
	assert.Contains(t, str, "<!-- coderank:coderank:end -->", "per-skill end marker must be present")
	assert.Contains(t, str, "# CodeRank - Library Documentation", "body must be written")
	assert.Contains(t, str, "ALWAYS run", "trigger instructions must be present")
	assert.NotContains(t, str, "name: coderank", "YAML frontmatter must be stripped for single-file agents")
	assert.NotContains(t, str, "description: Query current", "frontmatter description must be stripped")
}

func TestEmitSkillSingleMarkerFileCoexistRootAndWiki(t *testing.T) {
	// Regression guard for the bug introduced in 3c-iii and fixed in 3c-v:
	// before per-skill markers, a second EmitSkill call with a different
	// skillName overwrote the first because both used the same marker pair.
	// With per-skill markers, root and wiki sections must both survive.
	dir := t.TempDir()
	codex, _ := FindByIDs([]string{"codex"})

	rootBody := "---\nname: coderank\ndescription: d\n---\n\n# Root body\nroot text\n"
	wikiBody := "---\nname: coderank-wiki\ndescription: d\n---\n\n# Wiki body\nwiki text\n"

	require.NoError(t, EmitSkill(dir, codex[0], ScopeProject, "coderank", rootBody))
	require.NoError(t, EmitSkill(dir, codex[0], ScopeProject, "coderank-wiki", wikiBody))

	got, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	str := string(got)

	assert.Contains(t, str, "<!-- coderank:coderank:start -->")
	assert.Contains(t, str, "<!-- coderank:coderank:end -->")
	assert.Contains(t, str, "<!-- coderank:coderank-wiki:start -->")
	assert.Contains(t, str, "<!-- coderank:coderank-wiki:end -->")
	assert.Contains(t, str, "root text", "root body must survive after wiki is written")
	assert.Contains(t, str, "wiki text", "wiki body must be present")
}

func TestEmitSkillSingleMarkerFilePreservesExistingContent(t *testing.T) {
	dir := t.TempDir()
	agentsPath := filepath.Join(dir, "AGENTS.md")
	existing := "# My project rules\n\n- prefer TypeScript\n- run make test before commit\n"
	require.NoError(t, os.WriteFile(agentsPath, []byte(existing), 0644))

	codex, _ := FindByIDs([]string{"codex"})
	require.NoError(t, EmitSkill(dir, codex[0], ScopeProject, "coderank", sampleSkill))

	got, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	str := string(got)
	assert.Contains(t, str, "My project rules", "user content before marker must survive")
	assert.Contains(t, str, "prefer TypeScript", "user content must survive")
	assert.Contains(t, str, "<!-- coderank:coderank:start -->", "per-skill start marker appended")
	assert.Contains(t, str, "# CodeRank - Library Documentation", "coderank body written")
}

func TestEmitSkillCursorMDC(t *testing.T) {
	dir := t.TempDir()
	cursor, _ := FindByIDs([]string{"cursor"})
	require.NoError(t, EmitSkill(dir, cursor[0], ScopeProject, "coderank", sampleSkill))

	got, err := os.ReadFile(filepath.Join(dir, ".cursor/rules/coderank.mdc"))
	require.NoError(t, err)
	str := string(got)
	assert.True(t, strings.HasPrefix(str, "---\n"), "MDC must start with frontmatter")
	assert.Contains(t, str, "description: Query current library documentation via coderank CLI. ALWAYS invoke.",
		"description must be rebuilt from the skill's description field")
	assert.Contains(t, str, "alwaysApply: true",
		"MDC must declare alwaysApply: true so the rule loads in every chat")
	assert.Contains(t, str, "# CodeRank - Library Documentation", "body must follow frontmatter")
	assert.NotContains(t, str, "name: coderank", "original SKILL.md frontmatter name field must not appear")
}

func TestEmitSkillCursorUnsupportedAtUserScope(t *testing.T) {
	dir := t.TempDir()
	cursor, _ := FindByIDs([]string{"cursor"})
	err := EmitSkill(dir, cursor[0], ScopeUser, "coderank", sampleSkill)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported at this scope",
		"Cursor has no user-scope format; EmitSkill must return an informative error")
}

func TestEmitSkillRespectsWindsurfMaxChars(t *testing.T) {
	dir := t.TempDir()
	windsurf, _ := FindByIDs([]string{"windsurf"})
	// Build content whose stripped body is over the 6000-char cap.
	bigBody := strings.Repeat("x", 7000)
	bigSkill := "---\nname: big\ndescription: big\n---\n\n" + bigBody

	err := EmitSkill(dir, windsurf[0], ScopeUser, "coderank", bigSkill)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds 6000-char cap",
		"Windsurf cap must be enforced to avoid silent truncation")
}

func TestStripFrontmatterHandlesNoFrontmatter(t *testing.T) {
	assert.Equal(t, "plain body\n", stripFrontmatter("plain body\n"),
		"content without frontmatter must be returned unchanged")
}

func TestStripFrontmatterHandlesMalformedFrontmatter(t *testing.T) {
	malformed := "---\nname: x\n(no closing delimiter)\n"
	assert.Equal(t, malformed, stripFrontmatter(malformed),
		"missing closing delimiter leaves content unchanged")
}

func TestExtractDescriptionHandlesQuotedValue(t *testing.T) {
	content := `---
name: x
description: "quoted description"
---

body`
	assert.Equal(t, "quoted description", extractDescription(content))
}
