// Package agents defines the registry of AI coding agents that coderank can
// target, their per-scope file formats, and path resolution helpers. Both the
// install and init commands use this registry to decide where to write skill
// content and how to format it.
package agents

import (
	"os"
	"path/filepath"

	"github.com/coderank-dev/coderank/internal/agentdetect"
)

// Skill name constants used when installing the root and wiki skills.
const (
	RootSkillName = "coderank"
	WikiSkillName = "coderank-wiki"
)

// SkillFormat describes how an agent consumes coderank content at a given scope.
type SkillFormat int

const (
	// FormatNone means the agent is not supported at this scope.
	FormatNone SkillFormat = iota
	// FormatSkillMD is a folder-per-skill layout with SKILL.md + YAML frontmatter.
	// Used by Claude Code, OpenCode, Kiro, and Antigravity.
	FormatSkillMD
	// FormatSingleMarkerFile is a single markdown file where coderank-managed
	// content lives between <!-- coderank:start --> / <!-- coderank:end -->
	// markers. Non-coderank content in the file is preserved. Used by Codex
	// (AGENTS.md), Gemini CLI (GEMINI.md), Copilot (copilot-instructions.md),
	// and Windsurf (global_rules.md / .windsurfrules).
	FormatSingleMarkerFile
	// FormatCursorMDC is a dedicated .mdc file with YAML frontmatter placed in
	// .cursor/rules/. Used by Cursor at project scope only (Cursor has no
	// supported user-scope rules path).
	FormatCursorMDC
)

// Scope distinguishes user-level (home directory) from project-level installs.
type Scope int

const (
	// ScopeUser resolves paths against the user's home directory.
	ScopeUser Scope = iota
	// ScopeProject resolves paths against the project root.
	ScopeProject
)

// AgentScope captures how coderank writes to and detects an agent at one scope.
// An empty Indicator.Path means the agent cannot be auto-detected at this scope
// and must be targeted explicitly via --agents. An empty Path together with
// FormatNone means the agent is entirely unsupported at this scope.
type AgentScope struct {
	Format    SkillFormat
	Path      string
	Indicator agentdetect.Indicator
	// MaxChars caps the total size coderank may write to a single-file target.
	// Zero means no cap. Windsurf global_rules.md enforces a 6000-char limit.
	MaxChars int
}

// Agent represents an AI coding agent with its user- and project-scope behavior.
type Agent struct {
	Name    string
	ID      string
	User    AgentScope
	Project AgentScope
}

// KnownAgents is the verified registry of agents coderank supports. Paths are
// sourced from each agent's official docs; the TestKnownAgentsPaths regression
// test locks these values in place.
var KnownAgents = []Agent{
	{
		Name: "Claude Code", ID: "claude",
		User:    AgentScope{Format: FormatSkillMD, Path: ".claude/skills", Indicator: agentdetect.Indicator{Path: ".claude", IsDir: true}},
		Project: AgentScope{Format: FormatSkillMD, Path: ".claude/skills", Indicator: agentdetect.Indicator{Path: ".claude", IsDir: true}},
	},
	{
		Name: "OpenCode", ID: "opencode",
		User:    AgentScope{Format: FormatSkillMD, Path: ".config/opencode/skills", Indicator: agentdetect.Indicator{Path: ".config/opencode", IsDir: true}},
		Project: AgentScope{Format: FormatSkillMD, Path: ".opencode/skills", Indicator: agentdetect.Indicator{Path: ".opencode", IsDir: true}},
	},
	{
		Name: "Kiro", ID: "kiro",
		User:    AgentScope{Format: FormatSkillMD, Path: ".kiro/skills", Indicator: agentdetect.Indicator{Path: ".kiro", IsDir: true}},
		Project: AgentScope{Format: FormatSkillMD, Path: ".kiro/skills", Indicator: agentdetect.Indicator{Path: ".kiro", IsDir: true}},
	},
	{
		Name: "Antigravity", ID: "antigravity",
		User:    AgentScope{Format: FormatSkillMD, Path: ".gemini/antigravity/skills", Indicator: agentdetect.Indicator{Path: ".gemini/antigravity", IsDir: true}},
		Project: AgentScope{Format: FormatSkillMD, Path: ".agent/skills", Indicator: agentdetect.Indicator{Path: ".agent", IsDir: true}},
	},
	{
		Name: "OpenAI Codex", ID: "codex",
		User:    AgentScope{Format: FormatSingleMarkerFile, Path: ".codex/AGENTS.md", Indicator: agentdetect.Indicator{Path: ".codex", IsDir: true}},
		// Project scope has no reliable auto-detect indicator: AGENTS.md is read
		// by multiple agents and .codex/ rarely appears at project level. Users
		// must pass --agents codex explicitly at project scope.
		Project: AgentScope{Format: FormatSingleMarkerFile, Path: "AGENTS.md"},
	},
	{
		Name: "Gemini CLI", ID: "gemini",
		User:    AgentScope{Format: FormatSingleMarkerFile, Path: ".gemini/GEMINI.md", Indicator: agentdetect.Indicator{Path: ".gemini", IsDir: true}},
		Project: AgentScope{Format: FormatSingleMarkerFile, Path: "GEMINI.md", Indicator: agentdetect.Indicator{Path: "GEMINI.md", IsDir: false}},
	},
	{
		Name: "GitHub Copilot", ID: "copilot",
		User:    AgentScope{Format: FormatSingleMarkerFile, Path: ".copilot/copilot-instructions.md", Indicator: agentdetect.Indicator{Path: ".copilot", IsDir: true}},
		Project: AgentScope{Format: FormatSingleMarkerFile, Path: ".github/copilot-instructions.md", Indicator: agentdetect.Indicator{Path: ".github", IsDir: true}},
	},
	{
		Name: "Windsurf", ID: "windsurf",
		User:    AgentScope{Format: FormatSingleMarkerFile, Path: ".codeium/windsurf/memories/global_rules.md", Indicator: agentdetect.Indicator{Path: ".codeium/windsurf", IsDir: true}, MaxChars: 6000},
		Project: AgentScope{Format: FormatSingleMarkerFile, Path: ".windsurfrules", Indicator: agentdetect.Indicator{Path: ".windsurfrules", IsDir: false}},
	},
	{
		Name: "Cursor", ID: "cursor",
		// Cursor has no working user-level rules path; a ~/.cursor/rules/
		// feature request exists but is unimplemented as of April 2026.
		User:    AgentScope{Format: FormatNone},
		Project: AgentScope{Format: FormatCursorMDC, Path: ".cursor/rules", Indicator: agentdetect.Indicator{Path: ".cursor", IsDir: true}},
	},
}

// InjectContextPath returns the project-relative path where `coderank inject`
// writes always-on context for this agent. Returns "" when inject isn't
// supported for this agent. Only the agents that the inject command currently
// targets (Claude, Cursor, Codex, Windsurf, plus the Generic fallback) are
// mapped; other agents can be added as inject gains support.
func (a Agent) InjectContextPath() string {
	switch a.ID {
	case "claude":
		return ".claude/context/coderank-stack.md"
	case "cursor":
		return ".cursor/rules/coderank-stack.mdc"
	case "codex":
		return "AGENTS.md"
	case "windsurf":
		return ".windsurfrules"
	case "generic":
		return ".coderank/stack.md"
	}
	return ""
}

// GenericAgent returns a synthetic agent used by `coderank inject` as a
// fallback target when no known agent is detected. Its InjectContextPath
// points at .coderank/stack.md inside the project - a location the user can
// configure any unknown agent to read.
func GenericAgent() Agent {
	return Agent{Name: "Generic", ID: "generic"}
}

// ScopeAt returns the agent's configuration for the given scope.
func (a Agent) ScopeAt(scope Scope) AgentScope {
	if scope == ScopeUser {
		return a.User
	}
	return a.Project
}

// SupportsScope reports whether the agent is installable at the given scope.
func (a Agent) SupportsScope(scope Scope) bool {
	return a.ScopeAt(scope).Format != FormatNone
}

// HasAutoDetect reports whether the agent can be auto-detected at the given
// scope. Agents without an indicator (e.g. Codex at project scope) require an
// explicit --agents target.
func (a Agent) HasAutoDetect(scope Scope) bool {
	s := a.ScopeAt(scope)
	return s.Format != FormatNone && s.Indicator.Path != ""
}

// ScopeRoot returns the filesystem root for the given scope: the user's home
// directory for ScopeUser, projectRoot for ScopeProject.
func ScopeRoot(scope Scope, projectRoot string) (string, error) {
	if scope == ScopeUser {
		return os.UserHomeDir()
	}
	return projectRoot, nil
}

// Detect returns agents whose indicator path exists at root for the given scope.
// Agents without an auto-detect indicator at this scope are never returned;
// callers should surface NonDetectable(scope) to the user as a hint.
func Detect(root string, scope Scope) []Agent {
	var found []Agent
	for _, agent := range KnownAgents {
		if !agent.HasAutoDetect(scope) {
			continue
		}
		if agent.ScopeAt(scope).Indicator.IsPresent(root) {
			found = append(found, agent)
		}
	}
	return found
}

// NonDetectable returns agents that are supported at the given scope but have
// no auto-detect indicator. Callers use this to warn the user that these
// agents require an explicit --agents flag to target.
func NonDetectable(scope Scope) []Agent {
	var out []Agent
	for _, agent := range KnownAgents {
		s := agent.ScopeAt(scope)
		if s.Format == FormatNone {
			continue
		}
		if s.Indicator.Path == "" {
			out = append(out, agent)
		}
	}
	return out
}

// FindByIDs returns agents matching the given IDs. Unknown IDs are returned
// separately so callers can present a helpful error.
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

// SkillPath returns the absolute path for a FormatSkillMD SKILL.md file at the
// given scope. Returns "" when the agent's format at this scope is not
// FormatSkillMD - callers must filter via agent.ScopeAt(scope).Format before
// calling. Non-SkillMD formats are written by format-specific emitters.
func SkillPath(root string, agent Agent, skillName string, scope Scope) string {
	s := agent.ScopeAt(scope)
	if s.Format != FormatSkillMD {
		return ""
	}
	return filepath.Join(root, s.Path, skillName, "SKILL.md")
}
