// Package inject handles ambient context injection — writing condensed
// API surfaces into agent context files so documentation is pre-loaded
// before the agent even starts. This is CodeRank's key differentiator:
// proactive docs instead of reactive queries.
package inject

import (
	"fmt"
	"os"
	"path/filepath"
)

// Agent represents a detected coding agent and where to write its context.
type Agent struct {
	// Name is the human-readable agent name (e.g., "Claude Code", "Cursor").
	Name string

	// ContextPath is the file path (relative to project root) where
	// CodeRank writes the injected API surfaces.
	ContextPath string
}

// DetectAgents checks the project directory for known agent configurations
// and returns the list of agents found. If no agents are detected, returns
// a generic target that any agent can read by convention.
//
// Detection is based on directory/file presence:
//   - .claude/ directory → Claude Code
//   - .cursor/ directory → Cursor
//   - AGENTS.md file → Codex
//   - .windsurfrules file → Windsurf
func DetectAgents(projectDir string) []Agent {
	var agents []Agent

	checks := []struct {
		indicator   string // file or directory to check for
		isDir       bool
		name        string
		contextPath string
	}{
		{".claude", true, "Claude Code", ".claude/context/coderank-stack.md"},
		{".cursor", true, "Cursor", ".cursor/rules/coderank-stack.mdc"},
		{"AGENTS.md", false, "Codex", "AGENTS.md"},
		{".windsurfrules", false, "Windsurf", ".windsurfrules"},
	}

	for _, check := range checks {
		path := filepath.Join(projectDir, check.indicator)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if check.isDir && !info.IsDir() {
			continue
		}
		agents = append(agents, Agent{
			Name:        check.name,
			ContextPath: check.contextPath,
		})
	}

	// If no known agents detected, use a generic file that any agent
	// can be configured to read.
	if len(agents) == 0 {
		agents = append(agents, Agent{
			Name:        "Generic",
			ContextPath: "coderank-context.md",
		})
	}

	return agents
}

// TargetForAgent returns the Agent config for a specific agent name.
// Used when the user passes --target to override auto-detection.
func TargetForAgent(name string) (Agent, error) {
	targets := map[string]Agent{
		"claude":   {Name: "Claude Code", ContextPath: ".claude/context/coderank-stack.md"},
		"cursor":   {Name: "Cursor", ContextPath: ".cursor/rules/coderank-stack.mdc"},
		"codex":    {Name: "Codex", ContextPath: "AGENTS.md"},
		"windsurf": {Name: "Windsurf", ContextPath: ".windsurfrules"},
		"generic":  {Name: "Generic", ContextPath: "coderank-context.md"},
	}

	agent, ok := targets[name]
	if !ok {
		return Agent{}, fmt.Errorf(
			"unknown agent %q — supported: claude, cursor, codex, windsurf, generic", name,
		)
	}
	return agent, nil
}
