package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EmitSkill writes skill content for the agent at the given scope, dispatching
// on the agent's format. The caller passes the original SKILL.md content (with
// frontmatter); EmitSkill adapts it to whatever the target agent expects:
//
//   - FormatSkillMD: written verbatim to <root>/<Path>/<skillName>/SKILL.md.
//   - FormatSingleMarkerFile: frontmatter stripped, body written between
//     coderank markers in the agent's single context file (AGENTS.md, etc.).
//     Pre-existing non-coderank content in that file is preserved.
//   - FormatCursorMDC: frontmatter rewritten as Cursor MDC fields
//     (description, alwaysApply: true), body kept. Written to
//     <root>/<Path>/<skillName>.mdc.
//   - FormatNone: returns an error; this agent isn't supported at this scope.
//
// When the target's scope caps the written section size (e.g. Windsurf's
// 6000-char global_rules.md limit), content longer than the cap returns an
// error rather than silently truncating.
func EmitSkill(root string, agent Agent, scope Scope, skillName, content string) error {
	s := agent.ScopeAt(scope)
	switch s.Format {
	case FormatSkillMD:
		return emitSkillMD(root, s, skillName, content)
	case FormatSingleMarkerFile:
		return emitSingleMarkerFile(root, s, skillName, agent.Name, content)
	case FormatCursorMDC:
		return emitCursorMDC(root, s, skillName, content)
	case FormatNone:
		return fmt.Errorf("%s not supported at this scope", agent.Name)
	default:
		return fmt.Errorf("unknown format %d for %s", s.Format, agent.Name)
	}
}

func emitSkillMD(root string, s AgentScope, skillName, content string) error {
	path := filepath.Join(root, s.Path, skillName, "SKILL.md")
	return WriteSkill(path, content)
}

func emitSingleMarkerFile(root string, s AgentScope, skillName, agentName, content string) error {
	body := stripFrontmatter(content)
	if s.MaxChars > 0 && len(body) > s.MaxChars {
		return fmt.Errorf("content for %s (%d chars) exceeds %d-char cap", agentName, len(body), s.MaxChars)
	}
	path := filepath.Join(root, s.Path)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", path, err)
	}
	start, end := markersFor(skillName)
	return WriteMarkerSectionCustom(path, body, start, end)
}

// markersFor returns per-skill start/end markers so multiple coderank sections
// (e.g. coderank + coderank-wiki) can coexist in the same shared file without
// overwriting one another.
func markersFor(skillName string) (string, string) {
	return fmt.Sprintf("<!-- coderank:%s:start -->", skillName),
		fmt.Sprintf("<!-- coderank:%s:end -->", skillName)
}

func emitCursorMDC(root string, s AgentScope, skillName, content string) error {
	description := extractDescription(content)
	body := stripFrontmatter(content)
	out := fmt.Sprintf("---\ndescription: %s\nalwaysApply: true\n---\n\n%s", description, body)
	path := filepath.Join(root, s.Path, skillName+".mdc")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", path, err)
	}
	return os.WriteFile(path, []byte(out), 0644)
}

// stripFrontmatter removes a leading YAML frontmatter block ("---\n...\n---\n")
// and returns the body. If the content has no frontmatter or is malformed, the
// original string is returned unchanged.
func stripFrontmatter(content string) string {
	rest, ok := strings.CutPrefix(content, "---\n")
	if !ok {
		return content
	}
	_, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return content
	}
	return strings.TrimLeft(body, "\n")
}

// extractDescription returns the value of the frontmatter "description:" field,
// or an empty string if not found. Assumes the description is on one line
// (YAML flow style). Handles values with or without surrounding quotes.
func extractDescription(content string) string {
	rest, ok := strings.CutPrefix(content, "---\n")
	if !ok {
		return ""
	}
	fm, _, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return ""
	}
	for line := range strings.SplitSeq(fm, "\n") {
		val, ok := strings.CutPrefix(line, "description:")
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		val = strings.TrimPrefix(val, `"`)
		val = strings.TrimSuffix(val, `"`)
		val = strings.TrimPrefix(val, `'`)
		val = strings.TrimSuffix(val, `'`)
		return val
	}
	return ""
}
