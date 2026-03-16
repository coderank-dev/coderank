// Package cmd implements all coderank CLI commands using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd is the base command for the coderank CLI.
var rootCmd = &cobra.Command{
	Use:   "coderank",
	Short: "AI-optimized library docs for coding agents",
	Long: `CodeRank provides modular, token-efficient documentation for coding agents.
Fetch library docs, manage your stack config, and inject context — all from the CLI.

Get started:
  coderank init           Set up .coderank.yml for your project
  coderank fetch react    Fetch React documentation
  coderank inject         Inject docs into your agent's context`,
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
	viper.AutomaticEnv()

	// It's fine if no config file exists — first-time users won't have one
	viper.ReadInConfig()
}
