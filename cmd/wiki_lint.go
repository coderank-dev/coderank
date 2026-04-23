package cmd

import (
	"fmt"
	"os"

	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
)

var wikiLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Scan the wiki for missing, orphaned, or deprecated pages",
	Long: `Scans .coderank/wiki/ for structural issues:

  - missing:    listed in index.md but no page file on disk
  - orphaned:   page file on disk but not listed in index.md
  - deprecated: page frontmatter marks status: deprecated

A [LINT] entry is appended to log.md with a summary of issues found. The
command returns a non-zero exit code when issues are present so it can be
used in CI or as a Stop hook.`,
	RunE: runWikiLint,
}

func init() {
	wikiCmd.AddCommand(wikiLintCmd)
}

func runWikiLint(cmd *cobra.Command, args []string) error {
	m := newWikiManager()
	result, err := m.Lint()
	if err != nil {
		return err
	}

	if len(result.Issues) == 0 {
		fmt.Fprintf(os.Stderr, "%d pages scanned; no issues found.\n", result.PagesScanned)
		_ = m.AppendLog("LINT", fmt.Sprintf("%d pages, 0 issues", result.PagesScanned))
		fmt.Print(render.SuccessMsg("Wiki lint passed"))
		return nil
	}

	fmt.Fprintf(os.Stderr, "%d pages scanned; %d issue(s):\n\n", result.PagesScanned, len(result.Issues))
	for _, issue := range result.Issues {
		fmt.Fprintf(os.Stderr, "  [%s] %s\n", issue.Kind, issue.Message)
	}
	_ = m.AppendLog("LINT", fmt.Sprintf("%d pages, %d issues", result.PagesScanned, len(result.Issues)))

	fmt.Print(render.WarningMsg(fmt.Sprintf("%d issue(s) found", len(result.Issues))))
	return fmt.Errorf("%d lint issue(s)", len(result.Issues))
}
