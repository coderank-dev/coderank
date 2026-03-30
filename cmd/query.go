package cmd

import (
	"fmt"
	"strings"

	"github.com/coderank-dev/coderank/internal/api"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(topicCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(gotchasCmd)
	rootCmd.AddCommand(topicsCmd)

	searchCmd.Flags().IntP("max-tokens", "t", 10000, "Maximum tokens in response")
}

// topicCmd fetches a full topic file by name.
var topicCmd = &cobra.Command{
	Use:   "topic <lib> <topic>",
	Short: "Full topic content by name",
	Long: `Fetches the full condensed content for a specific topic file.
Use 'coderank topics <lib>' to see available topic names.

Examples:
  coderank topic react hooks
  coderank topic zod validation
  coderank topic express routing`,
	Args: cobra.ExactArgs(2),
	RunE: runTopic,
}

// searchCmd performs a semantic keyword search across a library's docs.
var searchCmd = &cobra.Command{
	Use:   "search <lib> <keyword>",
	Short: "Semantic search across a library's docs",
	Long: `Performs semantic search over a library's condensed documentation
and returns the top 3 most relevant results.

Examples:
  coderank search react "server components"
  coderank search prisma "database transactions"
  coderank search zod "async validation"`,
	Args: cobra.ExactArgs(2),
	RunE: runSearch,
}

// gotchasCmd fetches gotcha/pitfall sections for a specific API.
var gotchasCmd = &cobra.Command{
	Use:   "gotchas <lib> <api-name>",
	Short: "Common pitfalls and gotchas for an API",
	Long: `Fetches the gotchas, pitfalls, and common mistakes for a specific API.
Useful when you know what to build but want to avoid the sharp edges.

Examples:
  coderank gotchas react useEffect
  coderank gotchas prisma transactions
  coderank gotchas zod union`,
	Args: cobra.ExactArgs(2),
	RunE: runGotchas,
}

// topicsCmd lists available topics for a library.
var topicsCmd = &cobra.Command{
	Use:   "topics <lib>",
	Short: "List available topics for a library",
	Long: `Lists all available topic files for a library. Use these names
with 'coderank topic <lib> <topic>' to fetch full content.

Examples:
  coderank topics react
  coderank topics zod
  coderank topics express`,
	Args: cobra.ExactArgs(1),
	RunE: runTopics,
}

func runTopic(cmd *cobra.Command, args []string) error {
	lib, topic := args[0], args[1]

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	resp, err := client.Topic(lib, topic)
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	if IsRawMode() {
		fmt.Print(resp.Content)
		return nil
	}

	fmt.Print(render.DocHeader(resp.Library, resp.Version, resp.Topic, resp.Tokens, 0))
	rendered, err := render.RenderMarkdown(resp.Content)
	if err != nil {
		fmt.Print(resp.Content)
	} else {
		fmt.Print(rendered)
	}
	return nil
}

func runSearch(cmd *cobra.Command, args []string) error {
	lib, keyword := args[0], args[1]
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	resp, err := client.Query(api.QueryRequest{
		Q:         keyword,
		MaxTokens: maxTokens,
		Library:   lib,
	})
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	return printQueryResponse(resp)
}

func runGotchas(cmd *cobra.Command, args []string) error {
	lib, apiName := args[0], args[1]

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	resp, err := client.Query(api.QueryRequest{
		Q:         apiName + " gotchas pitfalls common mistakes edge cases",
		MaxTokens: 3000,
		Library:   lib,
	})
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	if len(resp.Results) == 0 {
		fmt.Print(render.WarningMsg("No gotchas found for " + apiName))
		return nil
	}

	return printQueryResponse(resp)
}

func runTopics(cmd *cobra.Command, args []string) error {
	lib := args[0]

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	resp, err := client.Topics(lib)
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	if IsRawMode() {
		fmt.Println(strings.Join(resp.Topics, "\n"))
		return nil
	}

	fmt.Print(render.DocHeader(resp.Library, resp.Version, "topics", 0, 0))
	for _, topic := range resp.Topics {
		fmt.Printf("  • %s\n", topic)
	}
	return nil
}

// printQueryResponse renders a QueryResponse in raw or styled mode.
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
