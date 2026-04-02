package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectFindsClaudeAndCursor(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".cursor"), 0755))

	found := Detect(dir)

	ids := make([]string, len(found))
	for i, a := range found {
		ids[i] = a.ID
	}
	assert.Contains(t, ids, "claude")
	assert.Contains(t, ids, "cursor")
	assert.NotContains(t, ids, "codex", "codex dir not created")
}

func TestDetectReturnsEmptyWhenNoAgents(t *testing.T) {
	dir := t.TempDir()
	assert.Empty(t, Detect(dir))
}

func TestFindByIDs(t *testing.T) {
	found, unknown := FindByIDs([]string{"claude", "cursor", "fakeagent"})
	assert.Len(t, found, 2)
	assert.Equal(t, "claude", found[0].ID)
	assert.Equal(t, "cursor", found[1].ID)
	assert.Equal(t, []string{"fakeagent"}, unknown)
}

func TestSkillPathProject(t *testing.T) {
	agent := Agent{SkillsDir: ".claude/skills"}
	path := SkillPath("/project", agent, "coderank", false)
	assert.Equal(t, "/project/.claude/skills/coderank/SKILL.md", path)
}

func TestSkillPathGlobal(t *testing.T) {
	agent := Agent{SkillsDir: ".cursor/skills"}
	path := SkillPath("/project", agent, "react", true)

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".cursor/skills/react/SKILL.md")
	assert.Equal(t, expected, path)
}

func TestDetectAllNineAgents(t *testing.T) {
	dir := t.TempDir()
	for _, agent := range KnownAgents {
		configDir := filepath.Dir(agent.SkillsDir)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, configDir), 0755))
	}

	found := Detect(dir)
	assert.Len(t, found, len(KnownAgents), "should detect all known agents")
}

func TestRootSkillMDContent(t *testing.T) {
	content := RootSkillMD()

	assert.True(t, strings.HasPrefix(content, "---\n"))
	assert.Contains(t, content, "name: coderank")
	assert.Contains(t, content, "allowed-tools: Bash(coderank *)")

	assert.Contains(t, content, "coderank query")
	assert.Contains(t, content, "coderank topic")
	assert.Contains(t, content, "coderank search")
	assert.Contains(t, content, "coderank gotchas")
	assert.Contains(t, content, "coderank topics")

	estimatedTokens := len(content) / 4
	assert.Less(t, estimatedTokens, 500, "root skill should be under 500 tokens")
}
