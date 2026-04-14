package cmd

import (
	"fmt"
	"os"

	"github.com/coderank-dev/coderank/internal/update"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update CodeRank to the latest version",
	Long:  "Downloads and installs the latest CodeRank release from GitHub.\nDetects Homebrew installs and suggests the appropriate upgrade command instead.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(os.Stderr, "CodeRank %s → checking for updates...\n", buildVersion.version)

		result := update.Check(buildVersion.version)
		if result == nil || !result.UpdateAvail {
			fmt.Fprintf(os.Stderr, "✓ Already up to date (%s)\n", buildVersion.version)
			return nil
		}

		fmt.Fprintf(os.Stderr, "New version available: %s → %s\n", result.CurrentVersion, result.LatestVersion)
		return update.SelfUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
