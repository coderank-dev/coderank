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
	rootCmd.AddCommand(gotchasCmd)
	rootCmd.AddCommand(topicsCmd)
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
	body := render.StripFrontmatter(resp.Content)
	rendered, err := render.RenderMarkdown(body)
	if err != nil {
		fmt.Print(body)
	} else {
		fmt.Print(rendered)
	}
	return nil
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
