package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectProjectFindsClaudeAndCursor(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".cursor"), 0755))

	found := Detect(dir, ScopeProject)

	ids := make([]string, len(found))
	for i, a := range found {
		ids[i] = a.ID
	}
	assert.Contains(t, ids, "claude")
	assert.Contains(t, ids, "cursor")
	assert.NotContains(t, ids, "codex", "codex has no project indicator")
}

func TestDetectReturnsEmptyWhenNoAgents(t *testing.T) {
	dir := t.TempDir()
	assert.Empty(t, Detect(dir, ScopeProject))
	assert.Empty(t, Detect(dir, ScopeUser))
}

func TestFindByIDs(t *testing.T) {
	found, unknown := FindByIDs([]string{"claude", "cursor", "fakeagent"})
	assert.Len(t, found, 2)
	assert.Equal(t, "claude", found[0].ID)
	assert.Equal(t, "cursor", found[1].ID)
	assert.Equal(t, []string{"fakeagent"}, unknown)
}

func TestSkillPathProject(t *testing.T) {
	found, _ := FindByIDs([]string{"claude"})
	path := SkillPath("/project", found[0], "coderank", ScopeProject)
	assert.Equal(t, "/project/.claude/skills/coderank/SKILL.md", path)
}

func TestSkillPathUser(t *testing.T) {
	found, _ := FindByIDs([]string{"claude"})
	path := SkillPath("/home/user", found[0], "react", ScopeUser)
	assert.Equal(t, "/home/user/.claude/skills/react/SKILL.md", path)
}

func TestSkillPathReturnsEmptyForNonSkillMDFormat(t *testing.T) {
	codex, _ := FindByIDs([]string{"codex"})
	assert.Empty(t, SkillPath("/project", codex[0], "coderank", ScopeProject),
		"codex uses FormatSingleMarkerFile, not FormatSkillMD")
	assert.Empty(t, SkillPath("/home/user", codex[0], "coderank", ScopeUser),
		"codex user scope is single file too")

	cursor, _ := FindByIDs([]string{"cursor"})
	assert.Empty(t, SkillPath("/home/user", cursor[0], "coderank", ScopeUser),
		"cursor has no user-scope support")
	assert.Empty(t, SkillPath("/project", cursor[0], "coderank", ScopeProject),
		"cursor project scope uses FormatCursorMDC, not FormatSkillMD")
}

func TestKnownAgentsPaths(t *testing.T) {
	// Regression guard: any change to verified agent paths must update this map.
	type spec struct {
		userFormat    SkillFormat
		projectFormat SkillFormat
		userPath      string
		projectPath   string
	}
	expected := map[string]spec{
		"claude":      {FormatSkillMD, FormatSkillMD, ".claude/skills", ".claude/skills"},
		"opencode":    {FormatSkillMD, FormatSkillMD, ".config/opencode/skills", ".opencode/skills"},
		"kiro":        {FormatSkillMD, FormatSkillMD, ".kiro/skills", ".kiro/skills"},
		"antigravity": {FormatSkillMD, FormatSkillMD, ".gemini/antigravity/skills", ".agent/skills"},
		"codex":       {FormatSingleMarkerFile, FormatSingleMarkerFile, ".codex/AGENTS.md", "AGENTS.md"},
		"gemini":      {FormatSingleMarkerFile, FormatSingleMarkerFile, ".gemini/GEMINI.md", "GEMINI.md"},
		"copilot":     {FormatSingleMarkerFile, FormatSingleMarkerFile, ".copilot/copilot-instructions.md", ".github/copilot-instructions.md"},
		"windsurf":    {FormatSingleMarkerFile, FormatSingleMarkerFile, ".codeium/windsurf/memories/global_rules.md", ".windsurfrules"},
		"cursor":      {FormatNone, FormatCursorMDC, "", ".cursor/rules"},
	}
	assert.Len(t, KnownAgents, len(expected), "unexpected number of agents")
	for _, agent := range KnownAgents {
		want, ok := expected[agent.ID]
		require.True(t, ok, "unexpected agent id %s", agent.ID)
		assert.Equal(t, want.userFormat, agent.User.Format, "%s user format", agent.ID)
		assert.Equal(t, want.projectFormat, agent.Project.Format, "%s project format", agent.ID)
		assert.Equal(t, want.userPath, agent.User.Path, "%s user path", agent.ID)
		assert.Equal(t, want.projectPath, agent.Project.Path, "%s project path", agent.ID)
	}
}

func TestDetectAllAgentsAtProjectScope(t *testing.T) {
	dir := t.TempDir()
	// Create every agent's project indicator (Codex has none, so it stays undetected).
	for _, agent := range KnownAgents {
		ind := agent.Project.Indicator
		if ind.Path == "" {
			continue
		}
		full := filepath.Join(dir, ind.Path)
		if ind.IsDir {
			require.NoError(t, os.MkdirAll(full, 0755))
			continue
		}
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0755))
		require.NoError(t, os.WriteFile(full, []byte("x"), 0644))
	}
	found := Detect(dir, ScopeProject)
	assert.Len(t, found, 8, "8 of 9 agents have project auto-detection; codex has none")
}

func TestDetectAllAgentsAtUserScope(t *testing.T) {
	dir := t.TempDir()
	for _, agent := range KnownAgents {
		ind := agent.User.Indicator
		if ind.Path == "" {
			continue
		}
		full := filepath.Join(dir, ind.Path)
		if ind.IsDir {
			require.NoError(t, os.MkdirAll(full, 0755))
			continue
		}
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0755))
		require.NoError(t, os.WriteFile(full, []byte("x"), 0644))
	}
	found := Detect(dir, ScopeUser)
	assert.Len(t, found, 8, "8 of 9 agents have user auto-detection; cursor has none")
}

func TestNonDetectableScopes(t *testing.T) {
	projectNon := NonDetectable(ScopeProject)
	projectIDs := make([]string, len(projectNon))
	for i, a := range projectNon {
		projectIDs[i] = a.ID
	}
	assert.Equal(t, []string{"codex"}, projectIDs, "only codex lacks a project indicator")

	userNon := NonDetectable(ScopeUser)
	assert.Empty(t, userNon, "all user-supported agents have indicators; cursor is excluded via FormatNone")
}

func TestScopeRoot(t *testing.T) {
	project, err := ScopeRoot(ScopeProject, "/some/project")
	require.NoError(t, err)
	assert.Equal(t, "/some/project", project)

	user, err := ScopeRoot(ScopeUser, "/ignored")
	require.NoError(t, err)
	home, _ := os.UserHomeDir()
	assert.Equal(t, home, user)
}

func TestInjectContextPath(t *testing.T) {
	cases := map[string]string{
		"claude":   ".claude/context/coderank-stack.md",
		"cursor":   ".cursor/rules/coderank-stack.mdc",
		"codex":    "AGENTS.md",
		"windsurf": ".windsurfrules",
		// Not yet wired for inject
		"opencode":    "",
		"kiro":        "",
		"antigravity": "",
		"gemini":      "",
		"copilot":     "",
	}
	for id, want := range cases {
		found, _ := FindByIDs([]string{id})
		require.Len(t, found, 1, "agent %s missing", id)
		assert.Equal(t, want, found[0].InjectContextPath(), "%s inject path", id)
	}
	assert.Equal(t, ".coderank/stack.md", GenericAgent().InjectContextPath(), "generic fallback path")
}

func TestInjectContextPathMatchesProjectPathForSingleFileAgents(t *testing.T) {
	// For SingleMarkerFile agents that inject supports, the inject path must
	// equal the project-scope path - they're the same file. Drift between the
	// two would let skill writes and inject writes target different files and
	// silently diverge.
	for _, id := range []string{"codex", "windsurf"} {
		found, _ := FindByIDs([]string{id})
		require.Len(t, found, 1, "agent %s missing", id)
		agent := found[0]
		assert.Equal(t, agent.Project.Path, agent.InjectContextPath(),
			"%s: InjectContextPath must match Project.Path for single-file agents", id)
	}
}

func TestRootSkillMDContent(t *testing.T) {
	content := RootSkillMD()

	// Frontmatter
	assert.True(t, strings.HasPrefix(content, "---\n"), "must start with YAML frontmatter")
	assert.Contains(t, content, "name: coderank")
	assert.Contains(t, content, "description:")
	assert.NotContains(t, content, "user-invocable: false", "root skill must be user-invocable")

	// Core query commands
	assert.Contains(t, content, "coderank query")
	assert.Contains(t, content, "coderank topic")
	assert.Contains(t, content, "coderank surface")
	assert.Contains(t, content, "coderank gotchas")
	assert.Contains(t, content, "coderank health")
	assert.Contains(t, content, "coderank compare")
	assert.Contains(t, content, "coderank install")

	estimatedTokens := len(content) / 4
	assert.Less(t, estimatedTokens, 2000, "root skill should be under 2000 tokens")
}
