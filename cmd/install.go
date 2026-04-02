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

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install CodeRank skill into detected AI coding agents",
	Long: `Detects AI coding agents in the current project and installs a CodeRank
skill so agents know how to query library documentation.

By default, installs one root-level skill that covers all 300+ libraries.
Use --with-surfaces to also install per-library skills with inline API surfaces.

Supported agents: Claude Code, GitHub Copilot, OpenAI Codex, Cursor,
Windsurf, Gemini CLI, Kiro, Antigravity, OpenCode.

Examples:
  coderank install                                     # Root skill → all detected agents
  coderank install --global                            # Root skill → global (all projects)
  coderank install --with-surfaces react,express       # Also add per-library API surfaces
  coderank install --agents claude,cursor              # Target specific agents only
  coderank install --all-agents                        # Target all 9 known agents
  coderank install --dry-run                           # Preview without writing files`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("global", false, "Install to global agent skill directories (~/.{agent}/skills/)")
	installCmd.Flags().StringSlice("agents", nil, "Target specific agents by ID (comma-separated)")
	installCmd.Flags().Bool("dry-run", false, "Show what would be installed without writing files")
	installCmd.Flags().Bool("all-agents", false, "Install to all 9 known agents regardless of detection")
	installCmd.Flags().StringSlice("with-surfaces", nil, "Also install per-library skills with inline API surfaces (fetched from API)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	global, _ := cmd.Flags().GetBool("global")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	allAgentsFlag, _ := cmd.Flags().GetBool("all-agents")
	agentIDs, _ := cmd.Flags().GetStringSlice("agents")
	withSurfaces, _ := cmd.Flags().GetStringSlice("with-surfaces")

	projectRoot, _ := os.Getwd()

	// Resolve target agents
	var targetAgents []agents.Agent
	switch {
	case allAgentsFlag:
		targetAgents = agents.KnownAgents
	case len(agentIDs) > 0:
		var unknown []string
		targetAgents, unknown = agents.FindByIDs(agentIDs)
		if len(unknown) > 0 {
			return fmt.Errorf("unknown agent(s): %s\nKnown IDs: claude, copilot, codex, cursor, windsurf, gemini, kiro, antigravity, opencode",
				strings.Join(unknown, ", "))
		}
	default:
		targetAgents = agents.Detect(projectRoot)
		if len(targetAgents) == 0 {
			fmt.Fprintln(os.Stderr, "No AI agents detected in current directory.")
			fmt.Fprintln(os.Stderr, "Use --agents <id> to specify agents, or --all-agents to target all.")
			fmt.Fprintln(os.Stderr, "Detection checks for: .claude/, .cursor/, .github/, .codex/, .windsurf/, etc.")
			return nil
		}
	}

	agentNames := make([]string, len(targetAgents))
	for i, a := range targetAgents {
		agentNames[i] = a.Name
	}
	scope := "project"
	if global {
		scope = "global"
	}

	// 1. Always install root skill
	fmt.Fprintf(os.Stderr, "Installing CodeRank skill → %s (%s)\n", strings.Join(agentNames, ", "), scope)

	rootContent := agents.RootSkillMD()
	installed := 0

	for _, agent := range targetAgents {
		path := agents.SkillPath(projectRoot, agent, "coderank", global)
		if dryRun {
			fmt.Fprintf(os.Stderr, "  [dry-run] coderank → %s\n", path)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", agent.ID, err)
			continue
		}
		if err := os.WriteFile(path, []byte(rootContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", agent.ID, err)
			continue
		}
		installed++
	}
	if !dryRun {
		fmt.Fprintln(os.Stderr, "  ✓ coderank (root skill)")
	}

	// 2. Optionally install per-library skills
	if len(withSurfaces) > 0 {
		fmt.Fprintf(os.Stderr, "\nInstalling %d per-library skill(s) with API surfaces...\n", len(withSurfaces))

		apiClient, err := api.NewClient(viper.GetString("api-url"))
		if err != nil {
			return fmt.Errorf("creating API client: %w", err)
		}

		for _, lib := range withSurfaces {
			skillContent, err := apiClient.FetchSkill(lib)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", lib, err)
				continue
			}
			for _, agent := range targetAgents {
				path := agents.SkillPath(projectRoot, agent, lib, global)
				if dryRun {
					fmt.Fprintf(os.Stderr, "  [dry-run] %s → %s\n", lib, path)
					continue
				}
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					fmt.Fprintf(os.Stderr, "  ✗ %s/%s: %v\n", agent.ID, lib, err)
					continue
				}
				if err := os.WriteFile(path, []byte(skillContent), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "  ✗ %s/%s: %v\n", agent.ID, lib, err)
					continue
				}
				installed++
			}
			if !dryRun {
				fmt.Fprintf(os.Stderr, "  ✓ %s\n", lib)
			}
		}
	}

	if !dryRun {
		fmt.Fprintf(os.Stderr, "\n✓ Installed %d skill file(s) across %d agent(s)\n", installed, len(targetAgents))
	}

	return nil
}
