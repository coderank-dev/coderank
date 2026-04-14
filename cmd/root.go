// Package cmd implements all coderank CLI commands using Cobra.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/coderank-dev/coderank/internal/update"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// updateCheckResult holds the result of the background version check goroutine.
// It is set by PersistentPreRun and read by PersistentPostRun.
var updateCheckResult *update.CheckResult

// updateCheckDone is closed by the background goroutine when the check finishes.
var updateCheckDone chan struct{}

// rootCmd is the base command for the coderank CLI.
//
// Background update check: PersistentPreRun starts a goroutine that checks
// GitHub Releases for a newer version. PersistentPostRun waits for it and
// prints a one-line notice to stderr if an update is available.
//
// NOTE: Cobra does NOT inherit PersistentPreRun/PersistentPostRun from a parent
// command into a subcommand that defines its own PersistentPreRun/PersistentPostRun.
// If a subcommand ever needs its own PersistentPreRun, use PreRunE instead —
// that only runs for the specific command and does not conflict with this hook.
// See: https://github.com/spf13/cobra/issues/216
var rootCmd = &cobra.Command{
	Use:   "coderank",
	Short: "AI-optimized library docs for coding agents",
	Long: `CodeRank provides modular, token-efficient documentation for coding agents.
Fetch library docs, manage your stack config, and inject context — all from the CLI.

Get started:
  coderank init              Set up .coderank.yml for your project
  coderank query react hooks Query React documentation
  coderank inject            Inject docs into your agent's context`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if os.Getenv("CODERANK_NO_UPDATE_CHECK") == "1" {
			return
		}
		skip, _ := cmd.Flags().GetBool("skip-update-check")
		if skip {
			return
		}
		updateCheckDone = make(chan struct{})
		go func() {
			defer close(updateCheckDone)
			updateCheckResult = update.Check(buildVersion.version)
		}()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if updateCheckDone == nil {
			return
		}
		<-updateCheckDone
		if notice := updateCheckResult.NoticeString(); notice != "" {
			fmt.Fprint(os.Stderr, notice)
		}
	},
}

// buildVersion holds version info injected at build time via ldflags.
var buildVersion = struct {
	version, commit, date string
}{"dev", "none", "unknown"}

// SetVersion receives build-time version info from main.go.
func SetVersion(version, commit, date string) {
	buildVersion.version = version
	buildVersion.commit = commit
	buildVersion.date = date
	rootCmd.Version = version
}

// Execute runs the root command. Called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags — every subcommand inherits these
	rootCmd.PersistentFlags().Bool("skip-update-check", false, "Disable automatic update check (also: CODERANK_NO_UPDATE_CHECK=1)")
	rootCmd.PersistentFlags().Bool("raw", false, "Output raw markdown without styling")
	rootCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose/debug output")
	rootCmd.PersistentFlags().Bool("offline", false, "Use local cache only, no API calls")

	// Bind flags to Viper
	viper.BindPFlag("raw", rootCmd.PersistentFlags().Lookup("raw"))
	viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("offline", rootCmd.PersistentFlags().Lookup("offline"))
}

// IsRawMode returns true when output should be plain markdown —
// either --raw was passed explicitly, or stdout is not a TTY (piped).
func IsRawMode() bool {
	return viper.GetBool("raw") || !isatty.IsTerminal(os.Stdout.Fd())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".coderank")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.coderank")
	viper.SetEnvPrefix("CODERANK")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// It's fine if no config file exists — first-time users won't have one
	viper.ReadInConfig()
}
