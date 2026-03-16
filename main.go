// Package main is the entry point for the coderank CLI.
// CodeRank provides AI-optimized library documentation for coding agents,
// delivering modular markdown bundles with 5-10x fewer tokens than alternatives.
package main

import "github.com/coderank-dev/coderank/cmd"

// Build-time variables injected via GoReleaser ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	cmd.Execute()
}
