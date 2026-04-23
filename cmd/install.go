package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/coderank-dev/coderank/internal/agents"
	"github.com/spf13/cobra"
)

// installCmd is the parent for all install subcommands. It has no RunE; Cobra
// prints help text automatically when invoked without a subcommand.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install skills into AI coding agents",
	Long: `Install skills into AI coding agents.

Subcommands:
  lib <lib> [lib...]   Install per-library skills for the given libraries
  harness              Install the root + wiki harness skills (default: global scope)

Run 'coderank install <subcommand> --help' for details on each.`,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

// printNonDetectableHint surfaces agents that are supported at this scope but
// require explicit --agents targeting (no auto-detect indicator). Shared by
// install lib and install harness.
func printNonDetectableHint(scope agents.Scope) {
	non := agents.NonDetectable(scope)
	if len(non) == 0 {
		return
	}
	ids := make([]string, len(non))
	for i, a := range non {
		ids[i] = a.ID
	}
	fmt.Fprintf(os.Stderr, "Note: not auto-detected at this scope (add --agents %s to include): %s\n",
		strings.Join(ids, ","), strings.Join(ids, ", "))
}
