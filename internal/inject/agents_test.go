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

	agents := DetectAgents(dir)
	assert.Len(t, agents, 1)
	assert.Equal(t, "Claude Code", agents[0].Name)
	assert.Equal(t, ".claude/context/coderank-stack.md", agents[0].ContextPath)
}

func TestDetectAgentsFindsMultiple(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".cursor"), 0755))

	agents := DetectAgents(dir)
	assert.Len(t, agents, 2,
		"should detect both Claude Code and Cursor when both directories exist")
}

func TestDetectAgentsFallsBackToGeneric(t *testing.T) {
	dir := t.TempDir()
	// No agent directories — should return generic

	agents := DetectAgents(dir)
	assert.Len(t, agents, 1)
	assert.Equal(t, "Generic", agents[0].Name,
		"should fall back to generic target when no known agents are detected")
}

func TestTargetForAgentReturnsCorrectPaths(t *testing.T) {
	agent, err := TargetForAgent("claude")
	require.NoError(t, err)
	assert.Equal(t, "Claude Code", agent.Name)

	agent, err = TargetForAgent("cursor")
	require.NoError(t, err)
	assert.Contains(t, agent.ContextPath, ".cursor")
}

func TestTargetForAgentRejectsUnknown(t *testing.T) {
	_, err := TargetForAgent("vscode-copilot")
	assert.ErrorContains(t, err, "unknown agent",
		"should reject agent names we don't have a target mapping for")
}
