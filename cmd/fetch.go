package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coderank-dev/coderank/internal/api"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fetchCmd queries the CodeRank API for condensed library documentation
// and renders it in the terminal. This is the command developers use most —
// it's the equivalent of "give me the docs for X."
var fetchCmd = &cobra.Command{
	Use:   "fetch <query>",
	Short: "Fetch condensed documentation for a library",
	Long: `Queries the CodeRank API for condensed library documentation and renders
it in the terminal with syntax-highlighted code blocks.

The query is a natural language description of what you need. CodeRank
finds the most relevant documentation files and assembles them within
your token budget.

Examples:
  coderank fetch "react hooks"
  coderank fetch "nextjs middleware authentication" --max-tokens 3000
  coderank fetch "prisma migrations" --library prisma
  coderank fetch "react hooks" --raw | pbcopy     # pipe to clipboard
  coderank fetch "react hooks" --json              # structured output`,
	Aliases: []string{"f", "get"},
	Args:    cobra.MinimumNArgs(1),
	RunE:    runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().IntP("max-tokens", "t", 5000, "Maximum tokens in response")
	fetchCmd.Flags().StringP("library", "l", "", "Restrict results to a specific library")
	viper.BindPFlag("max-tokens", fetchCmd.Flags().Lookup("max-tokens"))
}

func runFetch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")
	maxTokens := viper.GetInt("max-tokens")
	library, _ := cmd.Flags().GetString("library")
	jsonOut := viper.GetBool("json")

	// Offline mode — only cache, no API call
	if viper.GetBool("offline") {
		return fmt.Errorf("no cached docs for %q — run 'coderank cache' first", query)
	}

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	// Build config from .coderank.yml if present
	var config *api.QueryConfig
	preferred := viper.GetStringSlice("stack.preferred")
	blocked := viper.GetStringSlice("stack.blocked")
	if len(preferred) > 0 || len(blocked) > 0 {
		config = &api.QueryConfig{
			Preferred:        preferred,
			Blocked:          blocked,
			PreferTypeScript: viper.GetBool("context.prefer_typescript"),
		}
	}

	resp, err := client.Query(api.QueryRequest{
		Q:         query,
		MaxTokens: maxTokens,
		Library:   library,
		Config:    config,
	})
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	if len(resp.Results) == 0 {
		fmt.Print(render.WarningMsg("No documentation found for: " + query))
		return nil
	}

	// --json: structured output for programmatic use
	if jsonOut {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// --raw or piped: plain markdown to stdout, no chrome, no Glamour rendering.
	// Agents consume this: coderank fetch --raw | ...
	if IsRawMode() {
		for _, result := range resp.Results {
			fmt.Print(result.Content)
			fmt.Println()
		}
		return nil
	}

	// Default: rendered markdown with Lip Gloss chrome
	for _, result := range resp.Results {
		fmt.Print(render.DocHeader(
			result.Library, result.Version, result.Topic, result.Tokens,
		))
		rendered, err := render.RenderMarkdown(result.Content)
		if err != nil {
			fmt.Print(result.Content)
		} else {
			fmt.Print(rendered)
		}
	}
	fmt.Print(render.DocFooter(resp.TotalTokens, resp.QueryMs))

	return nil
}
