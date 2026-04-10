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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in via browser to authenticate the CLI",
	Long: `Opens your browser to sign in to coderank.ai and automatically saves
your API key to ~/.coderank/credentials.

For local development, use --web-url and --api-url to point at local servers:
  coderank login --web-url http://localhost:3000 --api-url http://localhost:8787`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().String("web-url", "", "Web app URL (default: https://coderank.ai)")
	loginCmd.Flags().String("api-url", "", "API URL override saved to config (default: https://api.coderank.ai)")
	viper.BindPFlag("web-url", loginCmd.Flags().Lookup("web-url"))
	viper.BindPFlag("api-url", loginCmd.Flags().Lookup("api-url"))
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	webURL := viper.GetString("web-url")
	if webURL == "" {
		webURL = "https://coderank.ai"
	}
	apiURL := viper.GetString("api-url")

	// Generate random state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("generating state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Bind to a random available port on loopback
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
				errCh <- fmt.Errorf("state mismatch — possible CSRF attack")
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
	render.InfoMsg("Opening browser to authenticate…")

	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "\nCould not open browser. Visit this URL manually:\n  %s\n\n", authURL)
	}

	render.InfoMsg("Waiting for authentication (timeout: 5 minutes)…")

	select {
	case key := <-keyCh:
		srv.Close()
		return saveLoginCredentials(key, apiURL)
	case err := <-errCh:
		srv.Close()
		return fmt.Errorf("authentication failed: %w", err)
	case <-time.After(5 * time.Minute):
		srv.Close()
		return fmt.Errorf("timed out — run `coderank login` to try again")
	}
}

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

func saveLoginCredentials(key, apiURL string) error {
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

	// Persist a custom api-url so subsequent commands use the right endpoint.
	// Write only the api-url key — not the full viper state — to avoid
	// polluting the global config with flag defaults.
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
