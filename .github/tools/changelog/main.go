// Command changelog generates a user-facing release changelog by aggregating
// conventional commits across the coderank, api, and pipeline repos.
//
// Usage:
//
//	changelog --version v0.2.0 --token $GITHUB_TOKEN --out ./dist
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	version := flag.String("version", "", "Release version, e.g. v0.2.0 (required)")
	token := flag.String("token", "", "GitHub token for API access")
	org := flag.String("org", "coderank-dev", "GitHub organization")
	repos := flag.String("repos", "coderank,api,pipeline", "Comma-separated repos to aggregate")
	outDir := flag.String("out", ".", "Directory to write output files")
	flag.Parse()

	if *version == "" {
		fmt.Fprintln(os.Stderr, "error: --version is required")
		flag.Usage()
		os.Exit(1)
	}

	// Allow token from env as fallback.
	tok := *token
	if tok == "" {
		tok = os.Getenv("GITHUB_TOKEN")
	}

	repoList := strings.Split(*repos, ",")

	var allEntries []Entry
	for _, repo := range repoList {
		repo = strings.TrimSpace(repo)
		entries, err := FetchRepoEntries(*org, repo, tok)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", repo, err)
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	release := BuildRelease(*version, time.Now().UTC(), allEntries)

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: create output dir: %v\n", err)
		os.Exit(1)
	}

	// Write markdown (for GitHub Release body).
	mdPath := filepath.Join(*outDir, fmt.Sprintf("changelog-%s.md", release.Version))
	if err := os.WriteFile(mdPath, []byte(release.ToMarkdown()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: write markdown: %v\n", err)
		os.Exit(1)
	}

	// Write MDX (for web changelog page — consumed by UOW_46).
	mdxPath := filepath.Join(*outDir, fmt.Sprintf("%s.mdx", release.Version))
	if err := os.WriteFile(mdxPath, []byte(release.ToMDX()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: write mdx: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("changelog %s: %d user-facing changes (%d libraries, %d fixes, %d breaking)\n",
		release.Version,
		release.Stats.TotalChanges,
		release.Stats.NewLibraries,
		release.Stats.BugFixes,
		release.Stats.BreakingChanges,
	)
	fmt.Printf("  → %s\n", mdPath)
	fmt.Printf("  → %s\n", mdxPath)
}
