package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coderank-dev/coderank/internal/api"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/coderank-dev/coderank/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// compareCmd displays a comparison table of libraries in a category.
// This is one of two commands that uses Bubble Tea (interactive TUI).
var compareCmd = &cobra.Command{
	Use:   "compare <category>",
	Short: "Compare libraries in a category",
	Long: `Displays a scrollable comparison table of libraries ranked by health
score in a category. Navigate with arrow keys, quit with q.

Examples:
  coderank compare "node orm"
  coderank compare "react state management"
  coderank compare "python web framework" --json`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCompare,
}

func init() {
	rootCmd.AddCommand(compareCmd)
	compareCmd.Flags().IntP("limit", "n", 10, "Maximum libraries to compare")
}

func runCompare(cmd *cobra.Command, args []string) error {
	category := strings.Join(args, " ")
	limit, _ := cmd.Flags().GetInt("limit")
	jsonOut := viper.GetBool("json")

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	resp, err := client.Compare(category, limit)
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	if len(resp.Libraries) == 0 {
		fmt.Print(render.WarningMsg("No libraries found for category: " + category))
		return nil
	}

	// --json: structured output
	if jsonOut {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// --raw or piped: markdown table
	if IsRawMode() {
		fmt.Printf("# Compare: %s\n\n", category)
		fmt.Println("| Library | Score | Maintenance | Security | Community | Sustainability |")
		fmt.Println("|---------|-------|-------------|----------|-----------|----------------|")
		for _, lib := range resp.Libraries {
			fmt.Printf("| %s | %d | %d | %d | %d | %d |\n",
				lib.Library, lib.HealthScore,
				lib.Breakdown["maintenance"],
				lib.Breakdown["security"],
				lib.Breakdown["community"],
				lib.Breakdown["sustainability"])
		}
		return nil
	}

	// Default: interactive Bubble Tea table
	model := tui.NewCompareModel(resp)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("compare TUI failed: %w", err)
	}
	return nil
}
