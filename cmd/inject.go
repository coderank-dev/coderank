package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/coderank-dev/coderank/internal/api"
	"github.com/coderank-dev/coderank/internal/inject"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// injectCmd writes API surfaces for your stack into agent context files.
// This makes documentation available to the agent before it even asks —
// zero latency, zero queries, zero MCP overhead.
var injectCmd = &cobra.Command{
	Use:   "inject [libraries...]",
	Short: "Pre-inject API surfaces into agent context files",
	Long: `Writes condensed API surfaces for your stack directly into agent memory
files so documentation is available before the agent starts working.

With no arguments, reads libraries from .coderank.yml preferred list.
With arguments, injects only the specified libraries.

Auto-detects agents (Claude Code, Cursor, Codex, Windsurf) by checking
for their config directories. Use --target to override.

Examples:
  coderank inject                           # from .coderank.yml
  coderank inject react nextjs prisma       # specific libraries
  coderank inject --target claude           # force Claude Code target
  coderank inject --target cursor           # force Cursor target
  coderank inject --watch                   # auto-refresh on dep changes`,
	Args: cobra.ArbitraryArgs,
	RunE: runInject,
}

func init() {
	rootCmd.AddCommand(injectCmd)
	injectCmd.Flags().String("target", "", "Force a specific agent target (claude, cursor, codex, windsurf, generic)")
	injectCmd.Flags().Bool("watch", false, "Watch dependency files and auto-refresh on changes")
}

func runInject(cmd *cobra.Command, args []string) error {
	projectDir := "."
	target, _ := cmd.Flags().GetString("target")
	watch, _ := cmd.Flags().GetBool("watch")

	// Determine which libraries to inject
	var libraries []string
	if len(args) > 0 {
		libraries = args
	} else {
		// Read from .coderank.yml preferred list
		libraries = viper.GetStringSlice("stack.preferred")
		if len(libraries) == 0 {
			return fmt.Errorf(
				"no libraries specified and no stack.preferred in .coderank.yml\n" +
					"Run 'coderank init' to configure your stack, or specify libraries: coderank inject react nextjs",
			)
		}
	}

	// Detect or override agent targets
	var agents []inject.Agent
	if target != "" {
		agent, err := inject.TargetForAgent(target)
		if err != nil {
			return err
		}
		agents = []inject.Agent{agent}
	} else {
		agents = inject.DetectAgents(projectDir)
	}

	// Fetch API surfaces from the API
	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	fmt.Println(render.Title.Render("Injecting stack API surfaces..."))
	fmt.Println()

	var surfaces []string
	var totalTokens int

	for _, lib := range libraries {
		result, err := client.Surface(lib)
		if err != nil {
			fmt.Printf("  %s %s: %s\n", render.Warning.Render("⚠"), lib, err)
			continue
		}
		surfaces = append(surfaces, result.Content)
		totalTokens += result.Tokens
		fmt.Printf("  %s@%s %s\n",
			lib, result.Version,
			render.Subtle.Render(fmt.Sprintf("(%d tokens)", result.Tokens)),
		)
	}

	if len(surfaces) == 0 {
		return fmt.Errorf("no API surfaces fetched — check your library names")
	}

	// Build the context content
	content := inject.ContextContent{
		Libraries:   libraries,
		Body:        strings.Join(surfaces, "\n\n"),
		TotalTokens: totalTokens,
	}

	// Warn if total tokens are getting large
	if totalTokens > 6000 {
		fmt.Printf("\n  %s Stack context is %d tokens (%d libraries).\n",
			render.Warning.Render("⚠"), totalTokens, len(libraries))
		fmt.Println("  Consider reducing with: coderank inject --max-tokens 5000")
	}

	// Write to each detected agent
	fmt.Println()
	for _, agent := range agents {
		if err := inject.WriteContext(projectDir, agent, content); err != nil {
			fmt.Printf("  %s %s: %s\n", render.Error.Render("✗"), agent.Name, err)
			continue
		}
		fmt.Printf("  %s %s → %s\n",
			render.Success.Render("✓"), agent.Name, agent.ContextPath)
	}

	fmt.Printf("\n%s\n", render.Subtle.Render(
		fmt.Sprintf("─── %d libraries · %d tokens · %d agents ───",
			len(libraries), totalTokens, len(agents)),
	))

	// Suggest gitignore entries for generated context files
	fmt.Println()
	fmt.Println(render.Subtle.Render(
		"Tip: Add injected files to .gitignore (they're generated, not source):"))
	for _, agent := range agents {
		// Only suggest for dedicated files, not shared ones (AGENTS.md, .windsurfrules)
		if agent.Name != "Codex" && agent.Name != "Windsurf" {
			fmt.Println(render.Subtle.Render(
				"  echo '" + agent.ContextPath + "' >> .gitignore"))
		}
	}

	// Watch mode
	if watch {
		return runWatch(projectDir, libraries, agents, client)
	}

	return nil
}

// runWatch starts the file watcher and re-injects on dependency changes.
// Runs until interrupted with Ctrl+C.
func runWatch(projectDir string, libraries []string, agents []inject.Agent, client *api.Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	return inject.Watch(ctx, projectDir, func() error {
		// Re-read libraries from config in case .coderank.yml changed
		currentLibs := viper.GetStringSlice("stack.preferred")
		if len(currentLibs) == 0 {
			currentLibs = libraries
		}

		fmt.Println(render.Subtle.Render("Re-injecting..."))

		var surfaces []string
		var totalTokens int

		for _, lib := range currentLibs {
			result, err := client.Surface(lib)
			if err != nil {
				fmt.Printf("  %s %s: %s\n", render.Warning.Render("⚠"), lib, err)
				continue
			}
			surfaces = append(surfaces, result.Content)
			totalTokens += result.Tokens
		}

		if len(surfaces) == 0 {
			return fmt.Errorf("no API surfaces fetched")
		}

		content := inject.ContextContent{
			Libraries:   currentLibs,
			Body:        strings.Join(surfaces, "\n\n"),
			TotalTokens: totalTokens,
		}

		for _, agent := range agents {
			if err := inject.WriteContext(projectDir, agent, content); err != nil {
				fmt.Printf("  %s %s: %s\n", render.Error.Render("✗"), agent.Name, err)
			}
		}

		fmt.Printf("  %s Re-injected %d libraries (%d tokens)\n",
			render.Success.Render("✓"), len(currentLibs), totalTokens)
		return nil
	})
}
