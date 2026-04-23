package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/coderank-dev/coderank/internal/api"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// gotchasCmd surfaces common pitfalls for a specific API in a library.
// Backed by /v1/query with a canned pitfall-focused prompt; no dedicated
// endpoint is required on the API side.
var gotchasCmd = &cobra.Command{
	Use:   "gotchas <library> <api>",
	Short: "Show common pitfalls for a specific API in a library",
	Long: `Retrieves documented pitfalls, gotchas, and common mistakes for a specific
API, function, hook, or method in a library. Uses the same semantic search
backend as 'coderank query' with a canned pitfall-focused prompt.

Examples:
  coderank gotchas react useEffect
  coderank gotchas prisma findMany
  coderank gotchas zod "z.object"
  coderank gotchas react useEffect --raw | pbcopy
  coderank gotchas prisma transaction --json`,
	Args: cobra.ExactArgs(2),
	RunE: runGotchas,
}

func init() {
	rootCmd.AddCommand(gotchasCmd)
}

func runGotchas(cmd *cobra.Command, args []string) error {
	library := args[0]
	apiName := args[1]

	maxTokens := viper.GetInt("context.max_tokens")
	if maxTokens == 0 {
		maxTokens = 500
	}

	if viper.GetBool("offline") {
		return fmt.Errorf("gotchas requires a live API call; not available in offline mode")
	}

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	q := fmt.Sprintf("common pitfalls, gotchas, and mistakes when using %s in %s", apiName, library)

	resp, err := client.Query(api.QueryRequest{
		Library:   library,
		Q:         q,
		MaxTokens: maxTokens,
	})
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	if viper.GetBool("json") {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(resp.Results) == 0 {
		fmt.Print(render.WarningMsg(fmt.Sprintf(
			"No gotchas found for %q in %q. Try: coderank query %s \"<your question>\"",
			apiName, library, library,
		)))
		return nil
	}

	return printQueryResponse(resp)
}
