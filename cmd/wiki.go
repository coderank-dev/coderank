package cmd

import (
	"fmt"
	"os"

	"github.com/coderank-dev/coderank/internal/render"
	"github.com/coderank-dev/coderank/internal/wiki"
	"github.com/spf13/cobra"
)

var wikiCmd = &cobra.Command{
	Use:   "wiki",
	Short: "Manage the project knowledge wiki at .coderank/wiki/",
	Long: `Manage the project knowledge wiki - short, file-pointing notes about how
this specific project uses third-party libraries.

Wiki pages are not library documentation (that's 'coderank query <lib>').
They record project-specific decisions, patterns, and gotchas, with inline
pointers to the actual source files. A good page is 3-8 lines plus a list
of --refs pointing at the relevant code.

Subcommands:
  ingest  Record a new or updated page (atomic write + index + log update)
  list    Show the index of wiki pages
  read    Print a specific page
  log     Print recent [INGEST] / [LINT] activity from log.md
  lint    Scan for missing, orphaned, or deprecated pages`,
}

var wikiListCmd = &cobra.Command{
	Use:   "list",
	Short: "List wiki pages grouped by library",
	RunE:  runWikiList,
}

var wikiReadCmd = &cobra.Command{
	Use:   "read <lib> <topic>",
	Short: "Print a wiki page",
	Args:  cobra.ExactArgs(2),
	RunE:  runWikiRead,
}

var wikiLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Print recent wiki activity",
	RunE:  runWikiLog,
}

func init() {
	rootCmd.AddCommand(wikiCmd)
	wikiCmd.AddCommand(wikiListCmd)
	wikiCmd.AddCommand(wikiReadCmd)
	wikiCmd.AddCommand(wikiLogCmd)
	wikiLogCmd.Flags().Int("tail", 20, "Number of recent entries to show (0 = all)")
}

func runWikiList(cmd *cobra.Command, args []string) error {
	m := newWikiManager()
	pages, err := m.List()
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		fmt.Fprintln(os.Stderr, "No wiki pages yet. Run 'coderank wiki ingest' to record one.")
		return nil
	}
	var currentLib string
	for _, p := range pages {
		if p.Library != currentLib {
			if currentLib != "" {
				fmt.Println()
			}
			fmt.Printf("## %s\n", p.Library)
			currentLib = p.Library
		}
		line := fmt.Sprintf("- %s", p.Topic)
		if p.Description != "" {
			line += " - " + p.Description
		}
		fmt.Println(line)
	}
	return nil
}

func runWikiRead(cmd *cobra.Command, args []string) error {
	m := newWikiManager()
	page, err := m.Read(args[0], args[1])
	if err != nil {
		return err
	}
	if IsRawMode() {
		fmt.Print(page.Body)
		return nil
	}
	fmt.Printf("# %s - %s\n", page.Library, page.Topic)
	if page.Description != "" {
		fmt.Printf("_%s_\n\n", page.Description)
	} else {
		fmt.Println()
	}
	if page.Status != "" && page.Status != "current" {
		fmt.Printf("Status: **%s**\n\n", page.Status)
	}
	if len(page.Refs) > 0 {
		fmt.Println("Refs:")
		for _, r := range page.Refs {
			fmt.Printf("  - %s\n", r)
		}
		fmt.Println()
	}
	rendered, err := render.RenderMarkdown(page.Body)
	if err != nil {
		fmt.Print(page.Body)
	} else {
		fmt.Print(rendered)
	}
	return nil
}

func runWikiLog(cmd *cobra.Command, args []string) error {
	tail, _ := cmd.Flags().GetInt("tail")
	m := newWikiManager()
	out, err := m.Log(tail)
	if err != nil {
		return err
	}
	if out == "" {
		fmt.Fprintln(os.Stderr, "Log is empty. Run 'coderank init' or 'coderank wiki ingest' to seed it.")
		return nil
	}
	fmt.Print(out)
	return nil
}

func newWikiManager() *wiki.Manager {
	root, _ := os.Getwd()
	return wiki.New(root)
}
