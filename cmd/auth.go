package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// authCmd validates and stores an API key for CLI authentication.
// The key is validated against the API health endpoint with auth,
// then written to ~/.coderank/credentials with 0600 permissions.
var authCmd = &cobra.Command{
	Use:   "auth <api-key>",
	Short: "Authenticate the CLI with your API key",
	Long: `Validates your API key against the CodeRank API and stores it locally.
Get your API key at https://coderank.ai/dashboard after signing up.

Example:
  coderank auth cr_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`,
	Args: cobra.ExactArgs(1),
	RunE: runAuth,
}

func init() {
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) error {
	key := strings.TrimSpace(args[0])

	if !strings.HasPrefix(key, "cr_sk_") {
		return fmt.Errorf("invalid API key format — keys start with cr_sk_")
	}

	// Validate by calling the API with this key
	baseURL := viper.GetString("api-url")
	if baseURL == "" {
		baseURL = "https://api.coderank.ai"
	}

	req, err := http.NewRequest("GET", baseURL+"/health", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach CodeRank API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		render.ErrorMsg("Invalid API key")
		return fmt.Errorf("the API key was rejected — check it at coderank.ai/dashboard")
	}

	// Store credentials
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	credDir := filepath.Join(home, ".coderank")
	if err := os.MkdirAll(credDir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	credPath := filepath.Join(credDir, "credentials")
	if err := os.WriteFile(credPath, []byte(key), 0600); err != nil {
		return fmt.Errorf("writing credentials: %w", err)
	}

	fmt.Print(render.SuccessMsg(fmt.Sprintf(
		"API key validated. Stored in %s", credPath,
	)))
	return nil
}
