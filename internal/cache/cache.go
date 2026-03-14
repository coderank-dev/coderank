// Package cache provides a local SQLite-backed cache for offline CLI access.
// Cached files are stored at ~/.coderank/cache/ — the SQLite database tracks
// metadata (library, version, topic, tokens) while the actual markdown files
// live on disk alongside the database.
package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// CachedFile represents a single cached documentation file.
type CachedFile struct {
	Library  string
	Version  string
	Topic    string
	Tokens   int
	FilePath string
}

// Manager handles the local SQLite cache at ~/.coderank/cache/.
type Manager struct {
	db      *sql.DB
	baseDir string
}

// NewManager opens (or creates) the cache database at ~/.coderank/cache/cache.db.
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	baseDir := filepath.Join(home, ".coderank", "cache")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	dbPath := filepath.Join(baseDir, "cache.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening cache database: %w", err)
	}

	// Create schema on first run
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cached_files (
			library     TEXT NOT NULL,
			version     TEXT NOT NULL,
			topic       TEXT NOT NULL,
			tokens      INTEGER,
			cached_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			file_path   TEXT NOT NULL,
			PRIMARY KEY (library, version, topic)
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("creating cache schema: %w", err)
	}

	return &Manager{db: db, baseDir: baseDir}, nil
}

// Close closes the database connection.
func (m *Manager) Close() error {
	return m.db.Close()
}

// Put stores a file in the cache — writes content to disk and records
// metadata in SQLite.
func (m *Manager) Put(library, version, topic string, tokens int, content []byte) error {
	dir := filepath.Join(m.baseDir, library, version)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache dir for %s/%s: %w", library, version, err)
	}

	filePath := filepath.Join(dir, topic+".md")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	_, err := m.db.Exec(
		`INSERT OR REPLACE INTO cached_files (library, version, topic, tokens, file_path)
		 VALUES (?, ?, ?, ?, ?)`,
		library, version, topic, tokens, filePath,
	)
	return err
}

// Search finds cached files matching a keyword query against library and topic names.
// Used in offline mode as a fallback when the API is unavailable.
func (m *Manager) Search(query string, maxResults int) ([]CachedFile, error) {
	keywords := strings.Fields(strings.ToLower(query))
	if len(keywords) == 0 {
		return nil, nil
	}

	// Build LIKE conditions for each keyword
	conditions := make([]string, len(keywords))
	args := make([]any, 0, len(keywords)*2+1)
	for i, kw := range keywords {
		conditions[i] = "(LOWER(library) LIKE ? OR LOWER(topic) LIKE ?)"
		pattern := "%" + kw + "%"
		args = append(args, pattern, pattern)
	}
	args = append(args, maxResults)

	querySql := fmt.Sprintf(
		`SELECT library, version, topic, tokens, file_path
		 FROM cached_files WHERE %s
		 ORDER BY cached_at DESC LIMIT ?`,
		strings.Join(conditions, " AND "),
	)

	rows, err := m.db.Query(querySql, args...)
	if err != nil {
		return nil, fmt.Errorf("searching cache: %w", err)
	}
	defer rows.Close()

	var results []CachedFile
	for rows.Next() {
		var f CachedFile
		if err := rows.Scan(&f.Library, &f.Version, &f.Topic, &f.Tokens, &f.FilePath); err != nil {
			return nil, err
		}
		results = append(results, f)
	}
	return results, nil
}

// Stats returns the total number of cached files and total tokens.
func (m *Manager) Stats() (fileCount int, totalTokens int, err error) {
	row := m.db.QueryRow("SELECT COUNT(*), COALESCE(SUM(tokens), 0) FROM cached_files")
	err = row.Scan(&fileCount, &totalTokens)
	return
}
