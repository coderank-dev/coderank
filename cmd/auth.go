package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// authCmd authenticates the CLI. With no argument, opens a browser OAuth flow
// against coderank.ai. With a key argument, validates the key against the API
// and stores it. In both cases credentials land in ~/.coderank/credentials.
var authCmd = &cobra.Command{
	Use:   "auth [api-key]",
	Short: "Authenticate the CLI (browser flow, or pass a key directly)",
	Long: `Authenticate the CLI with your CodeRank account.

With no argument, opens your browser to sign in at coderank.ai and saves the
returned API key automatically.

With an API key argument, validates the key against the CodeRank API and
stores it without launching a browser.

Credentials are written to ~/.coderank/credentials with 0600 permissions.
Get your API key at https://coderank.ai/dashboard after signing up.

Examples:
  coderank auth                                                # browser flow
  coderank auth cr_sk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6         # direct key
  coderank auth --web-url http://localhost:3000                # local dev (browser)
  coderank auth cr_sk_... --api-url http://localhost:8787      # direct key, local API`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAuth,
}

func init() {
	authCmd.Flags().String("web-url", "", "Web app URL (default: https://coderank.ai)")
	authCmd.Flags().String("api-url", "", "API URL override saved to config (default: https://api.coderank.ai)")
	viper.BindPFlag("web-url", authCmd.Flags().Lookup("web-url"))
	viper.BindPFlag("api-url", authCmd.Flags().Lookup("api-url"))
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		return runAuthWithKey(strings.TrimSpace(args[0]))
	}
	return runAuthBrowser()
}

// runAuthWithKey validates the provided API key and stores it locally.
func runAuthWithKey(key string) error {
	if !strings.HasPrefix(key, "cr_sk_") {
		return fmt.Errorf("invalid API key format - keys start with cr_sk_")
	}

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
		return fmt.Errorf("the API key was rejected - check it at coderank.ai/dashboard")
	}

	apiURL := viper.GetString("api-url")
	return saveCredentials(key, apiURL)
}

// runAuthBrowser runs the browser-based OAuth flow against coderank.ai.
func runAuthBrowser() error {
	webURL := viper.GetString("web-url")
	if webURL == "" {
		webURL = "https://coderank.ai"
	}
	apiURL := viper.GetString("api-url")

	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("generating state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	keyCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			if q.Get("state") != state {
				http.Error(w, "Invalid state", http.StatusBadRequest)
				errCh <- fmt.Errorf("state mismatch - possible CSRF attack")
				return
			}

			key := q.Get("key")
			if key == "" || !strings.HasPrefix(key, "cr_sk_") {
				http.Error(w, "Invalid key", http.StatusBadRequest)
				errCh <- fmt.Errorf("invalid or missing key in callback")
				return
			}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<!doctype html><html><body style="font-family:sans-serif;padding:48px;text-align:center">
<h2>Authenticated!</h2><p>You can close this tab and return to the terminal.</p>
</body></html>`)
			keyCh <- key
		}),
	}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := fmt.Sprintf("%s/cli-auth?port=%d&state=%s", webURL, port, url.QueryEscape(state))
	render.InfoMsg("Opening browser to authenticate...")

	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "\nCould not open browser. Visit this URL manually:\n  %s\n\n", authURL)
	}

	render.InfoMsg("Waiting for authentication (timeout: 5 minutes)...")

	select {
	case key := <-keyCh:
		srv.Close()
		return saveCredentials(key, apiURL)
	case err := <-errCh:
		srv.Close()
		return fmt.Errorf("authentication failed: %w", err)
	case <-time.After(5 * time.Minute):
		srv.Close()
		return fmt.Errorf("timed out - run `coderank auth` to try again")
	}
}

// openBrowser opens rawURL in the user's default browser.
func openBrowser(rawURL string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", rawURL).Start()
	case "linux":
		return exec.Command("xdg-open", rawURL).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", rawURL).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// saveCredentials writes the API key to ~/.coderank/credentials and, if apiURL
// is non-empty, persists it to ~/.coderank/.coderank.yml so later commands
// target the right endpoint. The config file is intentionally minimal to
// avoid polluting the global config with flag defaults.
func saveCredentials(key, apiURL string) error {
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

	if apiURL != "" {
		configPath := filepath.Join(credDir, ".coderank.yml")
		data := fmt.Sprintf("api-url: %s\n", apiURL)
		if err := os.WriteFile(configPath, []byte(data), 0600); err != nil {
			return fmt.Errorf("saving api-url to config: %w", err)
		}
	}

	fmt.Print(render.SuccessMsg(fmt.Sprintf("Authenticated! API key saved to %s", credPath)))
	return nil
}
