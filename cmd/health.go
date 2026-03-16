package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/coderank-dev/coderank/internal/api"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// healthCmd displays the health score breakdown for a library.
// Scores are color-coded: green (≥70), yellow (40–69), red (<40).
var healthCmd = &cobra.Command{
	Use:   "health <library>",
	Short: "Show health score breakdown for a library",
	Long: `Displays a library's health score with sub-scores for maintenance,
security, community, and sustainability. Helps you evaluate whether
a library is actively maintained and safe to use.

Examples:
  coderank health lodash
  coderank health react --json`,
	Args: cobra.ExactArgs(1),
	RunE: runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	library := args[0]
	jsonOut := viper.GetBool("json")

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	result, err := client.Health(library)
	if err != nil {
		render.ErrorMsg("%s", err.Error())
		return err
	}

	if jsonOut {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Print(render.HealthDisplay(result))
	return nil
}
