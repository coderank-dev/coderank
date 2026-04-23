package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/coderank-dev/coderank/internal/agents"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
)

var installHarnessCmd = &cobra.Command{
	Use:   "harness",
	Short: "Install the CodeRank harness skills (root + wiki)",
	Long: `Install the CodeRank harness skills - the root 'coderank' skill and the
'coderank-wiki' skill - into your AI coding agents.

The harness teaches every agent to:
  - query coderank before answering library questions (root skill)
  - consult and maintain the project wiki at .coderank/wiki/ (wiki skill)

Defaults to global scope (~/.<agent>/...), so the harness is loaded in every
project you open. Use --project to install at project scope instead (equivalent
to what 'coderank init' does during project setup).

Emission is format-dispatched:
  - SkillMD agents (Claude, OpenCode, Kiro, Antigravity): folder + SKILL.md
  - Single-file agents (Codex, Gemini, Copilot, Windsurf): marker-bracketed
    sections in the agent's context file, preserving existing content
  - Cursor (project scope only): .mdc rule files in .cursor/rules/

Note: Cursor has no supported user-scope rules path, so it's only installable
with --project.

Examples:
  coderank install harness                      # global, all detected agents
  coderank install harness --project            # project scope
  coderank install harness --agents claude,kiro # specific agents
  coderank install harness --no-wiki            # only root skill
  coderank install harness --dry-run            # preview paths`,
	RunE: runInstallHarness,
}

func init() {
	installCmd.AddCommand(installHarnessCmd)
	installHarnessCmd.Flags().Bool("project", false, "Install at project scope instead of globally")
	installHarnessCmd.Flags().StringSlice("agents", nil, "Target specific agents by ID (comma-separated)")
	installHarnessCmd.Flags().Bool("no-wiki", false, "Skip the wiki skill; only install the root skill")
	installHarnessCmd.Flags().Bool("all-agents", false, "Install to all known agents regardless of detection")
	installHarnessCmd.Flags().Bool("dry-run", false, "Show what would be installed without writing files")
}

func runInstallHarness(cmd *cobra.Command, args []string) error {
	project, _ := cmd.Flags().GetBool("project")
	noWiki, _ := cmd.Flags().GetBool("no-wiki")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	allAgentsFlag, _ := cmd.Flags().GetBool("all-agents")
	agentIDs, _ := cmd.Flags().GetStringSlice("agents")

	scope := agents.ScopeUser
	if project {
		scope = agents.ScopeProject
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

	scopeLabel := "global"
	if project {
		scopeLabel = "project"
	}
	agentNames := make([]string, len(targets))
	for i, a := range targets {
		agentNames[i] = a.Name
	}
	fmt.Fprintf(os.Stderr, "Installing harness skills -> %s (%s)\n", strings.Join(agentNames, ", "), scopeLabel)

	rootContent := agents.RootSkillMD()
	wikiContent := agents.WikiSkillMD()

	installed := 0
	for _, agent := range targets {
		s := agent.ScopeAt(scope)
		if s.Format == agents.FormatNone {
			fmt.Fprintf(os.Stderr, "  ⚠ %s: not supported at %s scope\n", agent.Name, scopeLabel)
			continue
		}
		if dryRun {
			fmt.Fprintf(os.Stderr, "  [dry-run] %s <- %s\n", agent.Name, agents.RootSkillName)
			if !noWiki {
				fmt.Fprintf(os.Stderr, "  [dry-run] %s <- %s\n", agent.Name, agents.WikiSkillName)
			}
			continue
		}
		if err := agents.EmitSkill(scopeRoot, agent, scope, agents.RootSkillName, rootContent); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s/%s: %v\n", agent.ID, agents.RootSkillName, err)
			continue
		}
		if !noWiki {
			if err := agents.EmitSkill(scopeRoot, agent, scope, agents.WikiSkillName, wikiContent); err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s/%s: %v\n", agent.ID, agents.WikiSkillName, err)
				continue
			}
		}
		fmt.Fprintf(os.Stderr, "  ✓ %s\n", agent.Name)
		installed++
	}

	if !dryRun {
		fmt.Print(render.SuccessMsg(fmt.Sprintf("Installed CodeRank harness on %d agent(s) at %s scope", installed, scopeLabel)))
	}
	return nil
}
