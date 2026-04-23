package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coderank-dev/coderank/internal/render"
	"github.com/coderank-dev/coderank/internal/wiki"
	"github.com/spf13/cobra"
)

var wikiIngestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Record or update a wiki page (atomic: page + index + log)",
	Long: `Atomically records a wiki page, updates .coderank/wiki/index.md, and appends
an [INGEST] entry to .coderank/wiki/log.md.

Wiki pages should be short: a one-sentence summary, bullet points of key
files via --refs, and any non-obvious gotcha. They point at code, not
replace library documentation.

Body input modes (exactly one required):
  --summary "..."       # short inline body
  --body-from-stdin     # read body from stdin (pipe longer content)
  --body-from-file P    # read body from file P

Examples:
  coderank wiki ingest --lib zod --topic auth-schemas \
    --summary "Discriminated union pattern for login/signup" \
    --refs src/schemas/auth.ts,src/api/auth.ts \
    --description "Auth request/response schemas with narrow types"

  echo "longer body..." | coderank wiki ingest --lib react --topic hooks --body-from-stdin

  coderank wiki ingest --lib prisma --topic migrations --body-from-file notes/prisma.md`,
	RunE: runWikiIngest,
}

func init() {
	wikiCmd.AddCommand(wikiIngestCmd)
	wikiIngestCmd.Flags().String("lib", "", "Library name (required)")
	wikiIngestCmd.Flags().String("topic", "", "Topic slug (required)")
	wikiIngestCmd.Flags().String("summary", "", "Inline body content (short summaries)")
	wikiIngestCmd.Flags().Bool("body-from-stdin", false, "Read body from stdin")
	wikiIngestCmd.Flags().String("body-from-file", "", "Read body from the given file path")
	wikiIngestCmd.Flags().String("description", "", "One-line description shown in the wiki index (defaults to the body's first non-header line)")
	wikiIngestCmd.Flags().String("status", "current", "Page status: current | deprecated | under-review")
	wikiIngestCmd.Flags().StringSlice("refs", nil, "Project file paths this page points at (comma-separated)")
	wikiIngestCmd.Flags().StringSlice("related", nil, "Related pages as lib/topic (comma-separated)")
	_ = wikiIngestCmd.MarkFlagRequired("lib")
	_ = wikiIngestCmd.MarkFlagRequired("topic")
}

// softBodyLimit is the advisory threshold above which we nudge users toward
// shorter pages + --refs. No hard enforcement.
const softBodyLimit = 800

func runWikiIngest(cmd *cobra.Command, args []string) error {
	lib, _ := cmd.Flags().GetString("lib")
	topic, _ := cmd.Flags().GetString("topic")
	summary, _ := cmd.Flags().GetString("summary")
	fromStdin, _ := cmd.Flags().GetBool("body-from-stdin")
	fromFile, _ := cmd.Flags().GetString("body-from-file")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	refs, _ := cmd.Flags().GetStringSlice("refs")
	related, _ := cmd.Flags().GetStringSlice("related")

	body, err := resolveBody(summary, fromStdin, fromFile)
	if err != nil {
		return err
	}

	if len(body) > softBodyLimit {
		fmt.Fprintf(os.Stderr,
			"tip: wiki pages are best kept short (<%d chars). Consider a brief summary with --refs pointing at the real code.\n",
			softBodyLimit)
	}

	m := newWikiManager()
	page, err := m.Ingest(lib, topic, body, wiki.IngestOpts{
		Status:      status,
		Related:     related,
		Refs:        refs,
		Description: description,
	})
	if err != nil {
		return err
	}
	fmt.Print(render.SuccessMsg(fmt.Sprintf(
		"Recorded %s/%s at .coderank/wiki/%s/%s.md", page.Library, page.Topic, page.Library, page.Topic,
	)))
	return nil
}

// resolveBody picks the one active body-input mode. Exactly one of the three
// must be supplied; mixing them is a user error.
func resolveBody(summary string, fromStdin bool, fromFile string) (string, error) {
	modes := 0
	if summary != "" {
		modes++
	}
	if fromStdin {
		modes++
	}
	if fromFile != "" {
		modes++
	}
	if modes == 0 {
		return "", fmt.Errorf("provide one of --summary, --body-from-stdin, or --body-from-file")
	}
	if modes > 1 {
		return "", fmt.Errorf("use exactly one of --summary, --body-from-stdin, or --body-from-file")
	}
	switch {
	case summary != "":
		return strings.TrimSpace(summary) + "\n", nil
	case fromStdin:
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return strings.TrimSpace(string(b)) + "\n", nil
	case fromFile != "":
		b, err := os.ReadFile(fromFile)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", fromFile, err)
		}
		return strings.TrimSpace(string(b)) + "\n", nil
	}
	return "", fmt.Errorf("unreachable")
}
