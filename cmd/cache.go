package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/coderank-dev/coderank/internal/api"
	"github.com/coderank-dev/coderank/internal/cache"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cacheCmd manages local documentation cache for offline use.
// Subcommand-like behavior via flags: --status, --clear, --from-config.
var cacheCmd = &cobra.Command{
	Use:   "cache [libraries...]",
	Short: "Manage local documentation cache",
	Long: `Downloads condensed documentation to ~/.coderank/cache/ for offline use.
Run 'coderank fetch --offline' to use cached docs without network.

Examples:
  coderank cache react nextjs prisma     # cache specific libraries
  coderank cache --from-config           # cache everything in .coderank.yml
  coderank cache --status                # show what's cached
  coderank cache --clear                 # delete all cached docs`,
	RunE: runCache,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.Flags().Bool("status", false, "Show cache statistics")
	cacheCmd.Flags().Bool("clear", false, "Clear all cached documentation")
	cacheCmd.Flags().Bool("from-config", false, "Cache libraries from .coderank.yml")
}

func runCache(cmd *cobra.Command, args []string) error {
	showStatus, _ := cmd.Flags().GetBool("status")
	clearCache, _ := cmd.Flags().GetBool("clear")
	fromConfig, _ := cmd.Flags().GetBool("from-config")
	jsonOut := viper.GetBool("json")

	// Validate args early — before opening cache — so tests don't hit the filesystem.
	if !showStatus && !clearCache && !fromConfig && len(args) == 0 {
		return fmt.Errorf("specify libraries to cache, or use --from-config / --status / --clear")
	}

	store, err := cache.NewManager()
	if err != nil {
		return fmt.Errorf("opening cache: %w", err)
	}
	defer store.Close()

	// --status: show cache statistics
	if showStatus {
		fileCount, totalTokens, err := store.Stats()
		if err != nil {
			return err
		}
		libs, err := store.Libraries()
		if err != nil {
			return err
		}
		if jsonOut {
			data, _ := json.MarshalIndent(map[string]any{
				"file_count":   fileCount,
				"total_tokens": totalTokens,
				"libraries":    libs,
			}, "", "  ")
			fmt.Println(string(data))
			return nil
		}
		fmt.Printf("%s\n", render.Title.Render("Cache Status"))
		fmt.Printf("  Files:     %d\n", fileCount)
		fmt.Printf("  Tokens:    %d\n", totalTokens)
		fmt.Printf("  Libraries: %s\n", strings.Join(libs, ", "))
		return nil
	}

	// --clear: delete all cached docs
	if clearCache {
		if err := store.Clear(); err != nil {
			return err
		}
		fmt.Print(render.SuccessMsg("Cache cleared"))
		return nil
	}

	// Determine libraries to cache
	var libraries []string
	if fromConfig {
		libraries = viper.GetStringSlice("stack.preferred")
		if len(libraries) == 0 {
			return fmt.Errorf("no preferred libraries in .coderank.yml — run 'coderank init' first")
		}
	} else {
		libraries = args
	}

	client, err := api.NewClient(viper.GetString("api-url"))
	if err != nil {
		return err
	}

	p := progress.New(progress.WithDefaultGradient())

	for i, lib := range libraries {
		fmt.Printf("\r%s %s", p.ViewAs(float64(i)/float64(len(libraries))), lib)

		result, err := client.Surface(lib)
		if err != nil {
			render.ErrorMsg("Failed to cache %s: %s", lib, err)
			continue
		}

		if err := store.Put(lib, result.Version, "_api-surface", result.Tokens, []byte(result.Content)); err != nil {
			render.ErrorMsg("Failed to store %s: %s", lib, err)
			continue
		}
	}

	fmt.Printf("\r%s\n", p.ViewAs(1.0))
	fmt.Print(render.SuccessMsg(fmt.Sprintf("Cached %d libraries", len(libraries))))
	return nil
}

// formatBytes converts bytes to a human-readable string.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
