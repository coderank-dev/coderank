// Package agents defines the registry of AI coding agents that support
// SKILL.md files and provides path resolution for skill installation.
package agents

import (
	"os"
	"path/filepath"

	"github.com/coderank-dev/coderank/internal/agentdetect"
)

// Skill name constants used when installing skills into agent directories.
const (
	RootSkillName = "coderank"
	WikiSkillName = "coderank-wiki"
)

// Agent represents an AI coding agent that supports SKILL.md files.
type Agent struct {
	Name      string // display name, e.g. "Claude Code"
	ID        string // short identifier for --agents flag, e.g. "claude"
	SkillsDir string // relative path from project root, e.g. ".claude/skills"
}

// KnownAgents is the registry of all supported agents.
var KnownAgents = []Agent{
	{Name: "Claude Code", ID: "claude", SkillsDir: ".claude/skills"},
	{Name: "GitHub Copilot", ID: "copilot", SkillsDir: ".github/skills"},
	{Name: "OpenAI Codex", ID: "codex", SkillsDir: ".codex/skills"},
	{Name: "Cursor", ID: "cursor", SkillsDir: ".cursor/skills"},
	{Name: "Windsurf", ID: "windsurf", SkillsDir: ".windsurf/skills"},
	{Name: "Gemini CLI", ID: "gemini", SkillsDir: ".gemini/skills"},
	{Name: "Kiro", ID: "kiro", SkillsDir: ".kiro/skills"},
	{Name: "Antigravity", ID: "antigravity", SkillsDir: ".antigravity/skills"},
	{Name: "OpenCode", ID: "opencode", SkillsDir: ".opencode/skills"},
}

// Detect returns agents whose config directories exist in projectRoot.
func Detect(projectRoot string) []Agent {
	var found []Agent
	for _, agent := range KnownAgents {
		configDir := filepath.Dir(agent.SkillsDir) // ".claude/skills" → ".claude"
		ind := agentdetect.Indicator{Path: configDir, IsDir: true}
		if ind.IsPresent(projectRoot) {
			found = append(found, agent)
		}
	}
	return found
}

// FindByIDs returns agents matching the given IDs. Unknown IDs are returned separately.
func FindByIDs(ids []string) ([]Agent, []string) {
	lookup := make(map[string]Agent, len(KnownAgents))
	for _, a := range KnownAgents {
		lookup[a.ID] = a
	}
	var found []Agent
	var unknown []string
	for _, id := range ids {
		if a, ok := lookup[id]; ok {
			found = append(found, a)
		} else {
			unknown = append(unknown, id)
		}
	}
	return found, unknown
}

// WriteSkill writes skill content to path, creating parent directories as needed.
func WriteSkill(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// SkillPath returns the absolute path for a SKILL.md file.
//
//	SkillPath("/project", agent, "coderank", false) → "/project/.claude/skills/coderank/SKILL.md"
//	SkillPath("/project", agent, "react",    true)  → "~/.claude/skills/react/SKILL.md"
func SkillPath(projectRoot string, agent Agent, skillName string, global bool) string {
	var base string
	if global {
		homeDir, _ := os.UserHomeDir()
		base = filepath.Join(homeDir, agent.SkillsDir)
	} else {
		base = filepath.Join(projectRoot, agent.SkillsDir)
	}
	return filepath.Join(base, skillName, "SKILL.md")
}
