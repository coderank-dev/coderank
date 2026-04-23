package inject

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectAgentsFindsClaudeCode(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0755))

	found := DetectAgents(dir)
	assert.Len(t, found, 1)
	assert.Equal(t, "Claude Code", found[0].Name)
	assert.Equal(t, ".claude/context/coderank-stack.md", found[0].InjectContextPath())
}

func TestDetectAgentsFindsMultiple(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".cursor"), 0755))

	found := DetectAgents(dir)
	assert.Len(t, found, 2,
		"should detect both Claude Code and Cursor when both directories exist")
}

func TestDetectAgentsFallsBackToGeneric(t *testing.T) {
	dir := t.TempDir()
	// No agent directories - should return generic

	found := DetectAgents(dir)
	assert.Len(t, found, 1)
	assert.Equal(t, "Generic", found[0].Name,
		"should fall back to generic target when no known agents are detected")
	assert.Equal(t, ".coderank/stack.md", found[0].InjectContextPath())
}

func TestTargetForAgentReturnsCorrectPaths(t *testing.T) {
	agent, err := TargetForAgent("claude")
	require.NoError(t, err)
	assert.Equal(t, "Claude Code", agent.Name)

	agent, err = TargetForAgent("cursor")
	require.NoError(t, err)
	assert.Contains(t, agent.InjectContextPath(), ".cursor")

	agent, err = TargetForAgent("generic")
	require.NoError(t, err)
	assert.Equal(t, "Generic", agent.Name)
}

func TestTargetForAgentRejectsUnknown(t *testing.T) {
	_, err := TargetForAgent("vscode-copilot")
	assert.ErrorContains(t, err, "unknown agent",
		"should reject agent names we don't have a target mapping for")
}

func TestTargetForAgentRejectsUnwiredAgent(t *testing.T) {
	// Known in KnownAgents, but inject has no context path for it yet.
	_, err := TargetForAgent("opencode")
	assert.ErrorContains(t, err, "no inject context path",
		"should reject agents that have no inject support configured")
}
