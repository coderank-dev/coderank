package inject

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchedFiles are the dependency files that trigger a re-injection
// when modified. Covers all major ecosystems.
var WatchedFiles = []string{
	"package.json",
	"package-lock.json",
	"go.mod",
	"requirements.txt",
	"pyproject.toml",
	"Cargo.toml",
	".coderank.yml",
}

// Watch monitors dependency files for changes and calls onChange when
// a modification is detected. Uses debouncing (500ms) to prevent rapid
// re-triggers when editors save multiple times.
//
// Blocks until the context is cancelled. Returns nil on clean shutdown.
func Watch(ctx context.Context, projectDir string, onChange func() error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}
	defer watcher.Close()

	// Watch each dependency file that exists
	watchCount := 0
	for _, file := range WatchedFiles {
		path := filepath.Join(projectDir, file)
		if _, err := os.Stat(path); err == nil {
			if err := watcher.Add(path); err == nil {
				watchCount++
			}
		}
	}

	if watchCount == 0 {
		return fmt.Errorf("no dependency files found to watch in %s", projectDir)
	}

	fmt.Printf("Watching %d dependency files for changes...\n", watchCount)

	// Debounce timer — prevents re-triggering within 500ms of last event
	var debounceTimer *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Reset debounce timer on each event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
				timestamp := time.Now().Format("15:04:05")
				fmt.Printf("[%s] Detected change in %s\n", timestamp, filepath.Base(event.Name))
				if err := onChange(); err != nil {
					fmt.Printf("[%s] Error: %s\n", timestamp, err)
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watcher error: %s\n", err)

		case <-ctx.Done():
			return nil
		}
	}
}
