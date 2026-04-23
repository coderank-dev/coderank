package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/coderank-dev/coderank/internal/agents"
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
	injectCmd.Flags().Bool("surface", false, "Append full API surface after each skill (higher token cost)")
}

func runInject(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")
	watch, _ := cmd.Flags().GetBool("watch")
	withSurface, _ := cmd.Flags().GetBool("surface")

	var libraries []string
	if len(args) > 0 {
		libraries = args
	} else {
		libraries = viper.GetStringSlice("stack.preferred")
		if len(libraries) == 0 {
			return fmt.Errorf(
				"no libraries specified and no stack.preferred in .coderank.yml\n" +
					"Run 'coderank init' to configure your stack, or specify libraries: coderank inject react nextjs",
			)
		}
	}

	return runInjectWith(".", libraries, target, watch, withSurface)
}

// runInjectWith executes an inject cycle against explicit libraries. Used by
// `coderank init --inject` (and the interactive wizard) to trigger inject
// without going through Cobra flag parsing a second time.
func runInjectWith(projectDir string, libraries []string, target string, watch, withSurface bool) error {
	// Detect or override agent targets
	var targets []agents.Agent
	if target != "" {
		agent, err := inject.TargetForAgent(target)
		if err != nil {
			return err
		}
		targets = []agents.Agent{agent}
	} else {
		targets = inject.DetectAgents(projectDir)
	}

	// Fetch API surfaces from the API
	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	fmt.Println(render.Title.Render("Injecting stack skills..."))
	fmt.Println()

	var sections []string
	var totalTokens int

	for _, lib := range libraries {
		skill, err := client.FetchSkill(lib)
		if err != nil {
			fmt.Printf("  %s %s: %s\n", render.Warning.Render("⚠"), lib, err)
			continue
		}
		entry := skill
		skillTokens := len(strings.Fields(skill)) * 4 / 3

		if withSurface {
			surface, err := client.Surface(lib)
			if err != nil {
				fmt.Printf("  %s %s surface: %s\n", render.Warning.Render("⚠"), lib, err)
			} else {
				entry += "\n\n" + surface.Content
				skillTokens += surface.Tokens
			}
		}

		sections = append(sections, entry)
		totalTokens += skillTokens
		fmt.Printf("  %s %s\n", lib,
			render.Subtle.Render(fmt.Sprintf("(%d tokens)", skillTokens)),
		)
	}

	if len(sections) == 0 {
		return fmt.Errorf("no skills fetched — check your library names")
	}

	// Build the context content
	content := inject.ContextContent{
		Libraries:   libraries,
		Body:        strings.Join(sections, "\n\n---\n\n"),
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
	for _, agent := range targets {
		if err := inject.WriteContext(projectDir, agent, content); err != nil {
			fmt.Printf("  %s %s: %s\n", render.Error.Render("✗"), agent.Name, err)
			continue
		}
		fmt.Printf("  %s %s -> %s\n",
			render.Success.Render("✓"), agent.Name, agent.InjectContextPath())
	}

	fmt.Printf("\n%s\n", render.Subtle.Render(
		fmt.Sprintf("--- %d libraries · %d tokens · %d agents ---",
			len(libraries), totalTokens, len(targets)),
	))

	// Suggest gitignore entries for generated context files
	fmt.Println()
	fmt.Println(render.Subtle.Render(
		"Tip: Add injected files to .gitignore (they're generated, not source):"))
	for _, agent := range targets {
		// Only suggest for dedicated files, not shared single-file agents.
		if agent.Project.Format != agents.FormatSingleMarkerFile {
			fmt.Println(render.Subtle.Render(
				"  echo '" + agent.InjectContextPath() + "' >> .gitignore"))
		}
	}

	// Watch mode
	if watch {
		return runWatch(projectDir, libraries, targets, client)
	}

	return nil
}

// runWatch starts the file watcher and re-injects on dependency changes.
// Runs until interrupted with Ctrl+C.
func runWatch(projectDir string, libraries []string, targets []agents.Agent, client *api.Client) error {
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

		var sections []string
		var totalTokens int

		for _, lib := range currentLibs {
			skill, err := client.FetchSkill(lib)
			if err != nil {
				fmt.Printf("  %s %s: %s\n", render.Warning.Render("⚠"), lib, err)
				continue
			}
			sections = append(sections, skill)
			totalTokens += len(strings.Fields(skill)) * 4 / 3
		}

		if len(sections) == 0 {
			return fmt.Errorf("no skills fetched")
		}

		content := inject.ContextContent{
			Libraries:   currentLibs,
			Body:        strings.Join(sections, "\n\n---\n\n"),
			TotalTokens: totalTokens,
		}

		for _, agent := range targets {
			if err := inject.WriteContext(projectDir, agent, content); err != nil {
				fmt.Printf("  %s %s: %s\n", render.Error.Render("✗"), agent.Name, err)
			}
		}

		fmt.Printf("  %s Re-injected %d libraries (%d tokens)\n",
			render.Success.Render("✓"), len(currentLibs), totalTokens)
		return nil
	})
}
