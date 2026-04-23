// Package inject handles ambient context injection - writing condensed
// API surfaces into agent context files so documentation is pre-loaded
// before the agent even starts. This is CodeRank's always-on alternative
// to skills: docs are in context from turn 1, not loaded on demand.
//
// The agent registry and marker-based file writer live in the agents
// package; this package provides the inject-specific glue: detection with
// a Generic fallback, --target string resolution, and context writing.
package inject

import (
	"fmt"

	"github.com/coderank-dev/coderank/internal/agents"
)

// DetectAgents returns agents that inject supports and whose indicators are
// present in projectDir. If none match, returns a single Generic agent so
// inject still has a target to write to.
func DetectAgents(projectDir string) []agents.Agent {
	detected := agents.Detect(projectDir, agents.ScopeProject)
	var withInject []agents.Agent
	for _, a := range detected {
		if a.InjectContextPath() != "" {
			withInject = append(withInject, a)
		}
	}
	if len(withInject) == 0 {
		return []agents.Agent{agents.GenericAgent()}
	}
	return withInject
}

// TargetForAgent resolves a --target flag string (e.g. "claude", "generic") to
// a concrete agent. Only agents that inject supports are accepted here; the
// Generic fallback is addressable via "generic".
func TargetForAgent(name string) (agents.Agent, error) {
	if name == "generic" {
		return agents.GenericAgent(), nil
	}
	found, unknown := agents.FindByIDs([]string{name})
	if len(unknown) > 0 || len(found) == 0 {
		return agents.Agent{}, fmt.Errorf(
			"unknown agent %q - supported inject targets: claude, cursor, codex, windsurf, generic", name,
		)
	}
	if found[0].InjectContextPath() == "" {
		return agents.Agent{}, fmt.Errorf(
			"%q has no inject context path configured - supported inject targets: claude, cursor, codex, windsurf, generic", name,
		)
	}
	return found[0], nil
}
