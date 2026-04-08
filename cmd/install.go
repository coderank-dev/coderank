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
	Use:   "install <lib> [lib...]",
	Short: "Install per-library API surface skills into detected AI coding agents",
	Long: `Fetches condensed API surface skills for the specified libraries and installs
them into detected AI coding agents in the current project.

Per-library skills provide inline API signatures so agents don't need to call
'coderank query' for every common operation.

Run 'coderank init' once per project to install the root CodeRank skill and
set up the project wiki.

Supported agents: Claude Code, GitHub Copilot, OpenAI Codex, Cursor,
Windsurf, Gemini CLI, Kiro, Antigravity, OpenCode.

Examples:
  coderank install react
  coderank install react express zod
  coderank install react --agents claude,cursor
  coderank install react --dry-run`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("global", false, "Install to global agent skill directories (~/.{agent}/skills/)")
	installCmd.Flags().StringSlice("agents", nil, "Target specific agents by ID (comma-separated)")
	installCmd.Flags().Bool("dry-run", false, "Show what would be installed without writing files")
	installCmd.Flags().Bool("all-agents", false, "Install to all 9 known agents regardless of detection")
}

func runInstall(cmd *cobra.Command, args []string) error {
	global, _ := cmd.Flags().GetBool("global")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	allAgentsFlag, _ := cmd.Flags().GetBool("all-agents")
	agentIDs, _ := cmd.Flags().GetStringSlice("agents")

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

	apiClient, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Installing %d library skill(s) → %s (%s)\n", len(args), strings.Join(agentNames, ", "), scope)

	installed := 0
	for _, lib := range args {
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

	if !dryRun {
		fmt.Fprintf(os.Stderr, "\n✓ Installed %d skill file(s) across %d agent(s)\n", installed, len(targetAgents))
	}

	return nil
}
