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

var queryCmd = &cobra.Command{
	Use:   "query <library> <question>",
	Short: "Query condensed documentation for a library",
	Long: `Queries the CodeRank API for condensed library documentation and renders
it in the terminal with syntax-highlighted code blocks.

The first argument is the library name — results are filtered to that library only.
The remaining arguments form the natural language question.

Examples:
  coderank query react "useCallback vs useMemo"
  coderank query prisma "how do migrations work" --max-tokens 3000
  coderank query requests "how to set Content-Type to JSON"
  coderank query react hooks --raw | pbcopy     # pipe to clipboard
  coderank query zod "object schema" --json      # structured output`,
	Args: cobra.MinimumNArgs(2),
	RunE: runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().IntP("max-tokens", "t", 500, "Maximum tokens in response")
	queryCmd.Flags().StringP("library", "l", "", "Restrict results to a specific library")
	viper.BindPFlag("max-tokens", queryCmd.Flags().Lookup("max-tokens"))
}

func runQuery(cmd *cobra.Command, args []string) error {
	library := args[0]
	query := strings.Join(args[1:], " ")
	maxTokens := viper.GetInt("max-tokens")
	// --library flag overrides positional library arg if explicitly set
	if flagLib, _ := cmd.Flags().GetString("library"); flagLib != "" {
		library = flagLib
	}
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
	// Agents consume this: coderank query --raw | ...
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
			result.Library, result.Version, result.Topic, result.Tokens, result.Score,
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

// printQueryResponse renders a QueryResponse in raw or styled mode.
// Shared by query subcommands (gotchas, search).
func printQueryResponse(resp *api.QueryResponse) error {
	if len(resp.Results) == 0 {
		fmt.Print(render.WarningMsg("No documentation found"))
		return nil
	}

	if IsRawMode() {
		for _, result := range resp.Results {
			fmt.Print(result.Content)
			fmt.Println()
		}
		return nil
	}

	for _, result := range resp.Results {
		if result.Content == "" {
			continue
		}
		fmt.Print(render.DocHeader(result.Library, result.Version, result.Topic, result.Tokens, result.Score))
		body := render.StripFrontmatter(result.Content)
		rendered, err := render.RenderMarkdown(body)
		if err != nil {
			fmt.Print(body)
		} else {
			fmt.Print(rendered)
		}
	}
	fmt.Print(render.DocFooter(resp.TotalTokens, resp.QueryMs))
	return nil
}
