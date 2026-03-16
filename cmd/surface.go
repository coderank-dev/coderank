package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/coderank-dev/coderank/internal/api"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// surfaceCmd fetches and displays the _api-surface.md file for a library.
// This is a lightweight call — no vector search, just a direct file lookup.
// The API surface is a compact listing of every public API (~1,500 tokens).
var surfaceCmd = &cobra.Command{
	Use:   "surface <library>",
	Short: "Show the API surface for a library",
	Long: `Displays a compact listing of every public API in a library — function
names, type signatures, one-line descriptions. ~1,500 tokens.

This gives you (or your agent) a complete map of what's available
without needing the full documentation.

Examples:
  coderank surface react
  coderank surface nextjs --raw
  coderank surface prisma --json`,
	Args: cobra.ExactArgs(1),
	RunE: runSurface,
}

func init() {
	rootCmd.AddCommand(surfaceCmd)
}

func runSurface(cmd *cobra.Command, args []string) error {
	library := args[0]
	jsonOut := viper.GetBool("json")

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	result, err := client.Surface(library)
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	if jsonOut {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if IsRawMode() {
		fmt.Print(result.Content)
		return nil
	}

	fmt.Print(render.DocHeader(result.Library, result.Version, "API Surface", result.Tokens))
	rendered, err := render.RenderMarkdown(result.Content)
	if err != nil {
		fmt.Print(result.Content)
	} else {
		fmt.Print(rendered)
	}
	return nil
}
