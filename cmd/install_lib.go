package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coderank-dev/coderank/internal/agents"
	"github.com/coderank-dev/coderank/internal/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installLibCmd = &cobra.Command{
	Use:   "lib <lib> [lib...]",
	Short: "Install per-library API surface skills into detected AI coding agents",
	Long: `Fetches condensed API surface skills for the specified libraries and installs
them into detected AI coding agents in the current project.

Per-library skills provide inline API signatures so agents don't need to call
'coderank query' for every common operation.

Run 'coderank init' once per project to install the root CodeRank skill and
set up the project wiki. Run 'coderank install harness' to install the
root + wiki skills globally for every project.

Supported formats: SkillMD (Claude Code, OpenCode, Kiro, Antigravity) and
CursorMDC (Cursor). Single-file agents (Codex, Gemini CLI, Copilot, Windsurf)
are not yet supported for per-library install; use 'coderank inject' for
always-on library context with those agents.

Note: Codex is not auto-detected at project scope. To target it explicitly,
pass --agents codex.

Examples:
  coderank install lib react
  coderank install lib react express zod
  coderank install lib react --agents claude,cursor
  coderank install lib react --dry-run`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInstallLib,
}

func init() {
	installCmd.AddCommand(installLibCmd)
	installLibCmd.Flags().Bool("global", false, "Install to user-level agent skill directories (~/)")
	installLibCmd.Flags().StringSlice("agents", nil, "Target specific agents by ID (comma-separated)")
	installLibCmd.Flags().Bool("dry-run", false, "Show what would be installed without writing files")
	installLibCmd.Flags().Bool("all-agents", false, "Install to all known agents regardless of detection")
}

func runInstallLib(cmd *cobra.Command, args []string) error {
	global, _ := cmd.Flags().GetBool("global")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	allAgentsFlag, _ := cmd.Flags().GetBool("all-agents")
	agentIDs, _ := cmd.Flags().GetStringSlice("agents")

	scope := agents.ScopeProject
	if global {
		scope = agents.ScopeUser
	}

	projectRoot, _ := os.Getwd()
	scopeRoot, err := agents.ScopeRoot(scope, projectRoot)
	if err != nil {
		return fmt.Errorf("resolving scope root: %w", err)
	}

	targets, err := resolveInstallTargets(scopeRoot, scope, allAgentsFlag, agentIDs)
	if err != nil {
		return err
	}
	if targets == nil {
		return nil
	}

	// Per-library install supports SkillMD and CursorMDC. SingleMarkerFile
	// agents would require per-library markers in a shared file - separate UX
	// decision - so they're filtered out here with a visible notice.
	var supported []agents.Agent
	for _, agent := range targets {
		f := agent.ScopeAt(scope).Format
		if f == agents.FormatSkillMD || f == agents.FormatCursorMDC {
			supported = append(supported, agent)
			continue
		}
		fmt.Fprintf(os.Stderr, "  ⚠ %s: per-library install not yet supported for this agent's format; use 'coderank inject'\n", agent.Name)
	}
	if len(supported) == 0 {
		fmt.Fprintln(os.Stderr, "No supported agents among the selected targets; nothing to install.")
		return nil
	}

	agentNames := make([]string, len(supported))
	for i, a := range supported {
		agentNames[i] = a.Name
	}
	scopeLabel := "project"
	if global {
		scopeLabel = "global"
	}

	apiClient, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Installing %d library skill(s) -> %s (%s)\n", len(args), strings.Join(agentNames, ", "), scopeLabel)

	installed := 0
	for _, lib := range args {
		skillContent, err := apiClient.FetchSkill(lib)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", lib, err)
			continue
		}

		surface, err := apiClient.Surface(lib)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s (surface): %v\n", lib, err)
			continue
		}

		for _, agent := range supported {
			if dryRun {
				skillPath := agents.SkillPath(scopeRoot, agent, lib, scope)
				if skillPath == "" {
					// CursorMDC path (not covered by SkillPath helper)
					skillPath = filepath.Join(scopeRoot, agent.ScopeAt(scope).Path, lib+".mdc")
				}
				fmt.Fprintf(os.Stderr, "  [dry-run] %s -> %s\n", lib, skillPath)
				continue
			}
			if err := agents.EmitSkill(scopeRoot, agent, scope, lib, skillContent); err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s/%s: %v\n", agent.ID, lib, err)
				continue
			}
			// For SkillMD agents, the surface file lives alongside SKILL.md.
			// CursorMDC has no separate surface file convention.
			if agent.ScopeAt(scope).Format == agents.FormatSkillMD {
				skillPath := agents.SkillPath(scopeRoot, agent, lib, scope)
				surfacePath := strings.TrimSuffix(skillPath, "SKILL.md") + "api-surface.md"
				if err := agents.WriteSkill(surfacePath, surface.Content); err != nil {
					fmt.Fprintf(os.Stderr, "  ✗ %s/%s api-surface.md: %v\n", agent.ID, lib, err)
					continue
				}
			}
			installed++
		}
		if !dryRun {
			fmt.Fprintf(os.Stderr, "  ✓ %s\n", lib)
		}
	}

	if !dryRun {
		fmt.Fprintf(os.Stderr, "\n✓ Installed %d skill file(s) across %d agent(s)\n", installed, len(supported))
	}
	return nil
}

// resolveInstallTargets picks the list of agents to act on based on --agents /
// --all-agents / auto-detection. Returns (nil, nil) with a user-facing message
// when nothing was detected and no explicit targets were provided.
func resolveInstallTargets(scopeRoot string, scope agents.Scope, allAgents bool, agentIDs []string) ([]agents.Agent, error) {
	switch {
	case allAgents:
		return agents.KnownAgents, nil
	case len(agentIDs) > 0:
		found, unknown := agents.FindByIDs(agentIDs)
		if len(unknown) > 0 {
			return nil, fmt.Errorf("unknown agent(s): %s\nKnown IDs: claude, opencode, kiro, antigravity, codex, gemini, copilot, windsurf, cursor",
				strings.Join(unknown, ", "))
		}
		return found, nil
	default:
		detected := agents.Detect(scopeRoot, scope)
		if len(detected) == 0 {
			fmt.Fprintln(os.Stderr, "No AI agents auto-detected in current scope.")
			fmt.Fprintln(os.Stderr, "Use --agents <id> to specify agents, or --all-agents to target all.")
			printNonDetectableHint(scope)
			return nil, nil
		}
		printNonDetectableHint(scope)
		return detected, nil
	}
}
