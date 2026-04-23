// Package wiki provides the read/write/lint operations for the project wiki
// at .coderank/wiki/. It backs the `coderank wiki` CLI commands and is also
// invoked by Claude Code hooks installed by `coderank init` so that wiki
// maintenance is a single deterministic tool call rather than a prose ritual
// the agent might skip.
package wiki

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// WikiDir is the project-relative location of the wiki.
const WikiDir = ".coderank/wiki"

// Page is the in-memory representation of a wiki page.
type Page struct {
	Library     string
	Topic       string
	Status      string   // "current" | "deprecated" | "under-review"
	Updated     string   // YYYY-MM-DD
	Related     []string // other pages as "lib/topic"
	Refs        []string // project file paths this page points at
	Description string   // one-line summary for the index
	Body        string   // free-form markdown body
}

// PageRef is a lightweight listing reference.
type PageRef struct {
	Library     string
	Topic       string
	Path        string // relative to wiki root
	Description string
}

// IngestOpts carries optional fields for Manager.Ingest.
type IngestOpts struct {
	Status      string
	Related     []string
	Refs        []string
	Description string
}

// LintIssue is a single problem surfaced by Lint.
type LintIssue struct {
	Kind    string // "missing" | "orphaned" | "deprecated" | "broken-ref"
	Page    PageRef
	Message string
}

// LintResult is the outcome of a Lint run.
type LintResult struct {
	Issues       []LintIssue
	PagesScanned int
	IndexEntries int
}

// Manager handles wiki operations rooted at projectRoot/.coderank/wiki/.
type Manager struct {
	Root string
}

// New constructs a Manager bound to the given project root.
func New(projectRoot string) *Manager {
	return &Manager{Root: filepath.Join(projectRoot, WikiDir)}
}

// Ingest atomically writes a wiki page, updates index.md, and appends to log.md.
// Re-ingesting the same lib/topic replaces the page and updates the index entry
// in place (no duplicate); the log gets a new [INGEST] entry every time.
func (m *Manager) Ingest(lib, topic, body string, opts IngestOpts) (Page, error) {
	lib = slug(lib)
	topic = slug(topic)
	if lib == "" || topic == "" {
		return Page{}, fmt.Errorf("lib and topic are required")
	}
	if err := os.MkdirAll(m.Root, 0755); err != nil {
		return Page{}, fmt.Errorf("creating wiki root: %w", err)
	}
	status := opts.Status
	if status == "" {
		status = "current"
	}
	desc := strings.TrimSpace(opts.Description)
	if desc == "" {
		desc = extractFirstDescLine(body)
	}
	page := Page{
		Library:     lib,
		Topic:       topic,
		Status:      status,
		Updated:     time.Now().Format("2006-01-02"),
		Related:     opts.Related,
		Refs:        opts.Refs,
		Description: desc,
		Body:        body,
	}
	pageDir := filepath.Join(m.Root, lib)
	if err := os.MkdirAll(pageDir, 0755); err != nil {
		return page, fmt.Errorf("creating lib directory: %w", err)
	}
	pagePath := filepath.Join(pageDir, topic+".md")
	if err := os.WriteFile(pagePath, []byte(renderPage(page)), 0644); err != nil {
		return page, fmt.Errorf("writing page: %w", err)
	}
	if err := m.updateIndex(page); err != nil {
		return page, fmt.Errorf("updating index: %w", err)
	}
	if err := m.AppendLog("INGEST", fmt.Sprintf("%s - %s", lib, topic)); err != nil {
		return page, fmt.Errorf("appending log: %w", err)
	}
	return page, nil
}

// Read returns the page at lib/topic, or a helpful error if absent.
func (m *Manager) Read(lib, topic string) (*Page, error) {
	lib = slug(lib)
	topic = slug(topic)
	path := filepath.Join(m.Root, lib, topic+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no page for %s/%s; try `coderank wiki list`", lib, topic)
		}
		return nil, err
	}
	page := parsePage(string(content))
	page.Library = lib
	page.Topic = topic
	return &page, nil
}

// List walks the wiki root and returns all pages sorted by library then topic.
func (m *Manager) List() ([]PageRef, error) {
	var refs []PageRef
	if _, err := os.Stat(m.Root); os.IsNotExist(err) {
		return refs, nil
	}
	entries, err := os.ReadDir(m.Root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		lib := entry.Name()
		subentries, err := os.ReadDir(filepath.Join(m.Root, lib))
		if err != nil {
			continue
		}
		for _, sub := range subentries {
			if sub.IsDir() {
				continue
			}
			name := sub.Name()
			if !strings.HasSuffix(name, ".md") {
				continue
			}
			topic := strings.TrimSuffix(name, ".md")
			desc := ""
			if content, err := os.ReadFile(filepath.Join(m.Root, lib, name)); err == nil {
				p := parsePage(string(content))
				desc = p.Description
			}
			refs = append(refs, PageRef{
				Library:     lib,
				Topic:       topic,
				Path:        filepath.Join(lib, name),
				Description: desc,
			})
		}
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Library != refs[j].Library {
			return refs[i].Library < refs[j].Library
		}
		return refs[i].Topic < refs[j].Topic
	})
	return refs, nil
}

// Log returns the contents of log.md, optionally limited to the last `tail`
// entries (lines starting with "["). Tail <= 0 returns the full log.
func (m *Manager) Log(tail int) (string, error) {
	logPath := filepath.Join(m.Root, "log.md")
	content, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if tail <= 0 {
		return string(content), nil
	}
	lines := strings.Split(string(content), "\n")
	var entryLines []int
	for i, line := range lines {
		if strings.HasPrefix(line, "[") {
			entryLines = append(entryLines, i)
		}
	}
	if len(entryLines) <= tail {
		return string(content), nil
	}
	startLine := entryLines[len(entryLines)-tail]
	return strings.Join(lines[startLine:], "\n"), nil
}

// Lint scans for missing, orphaned, and deprecated pages and returns the result.
func (m *Manager) Lint() (*LintResult, error) {
	result := &LintResult{}
	if _, err := os.Stat(m.Root); os.IsNotExist(err) {
		return result, nil
	}
	onDisk, err := m.List()
	if err != nil {
		return nil, err
	}
	result.PagesScanned = len(onDisk)
	diskKeys := map[string]PageRef{}
	for _, r := range onDisk {
		diskKeys[r.Library+"/"+r.Topic] = r
	}
	indexed, err := m.parseIndexEntries()
	if err != nil {
		return nil, err
	}
	result.IndexEntries = len(indexed)
	indexedKeys := map[string]bool{}
	for _, e := range indexed {
		key := e.Library + "/" + e.Topic
		indexedKeys[key] = true
		if _, ok := diskKeys[key]; !ok {
			result.Issues = append(result.Issues, LintIssue{
				Kind:    "missing",
				Page:    e,
				Message: fmt.Sprintf("listed in index but not on disk: %s/%s", e.Library, e.Topic),
			})
		}
	}
	for key, ref := range diskKeys {
		if !indexedKeys[key] {
			result.Issues = append(result.Issues, LintIssue{
				Kind:    "orphaned",
				Page:    ref,
				Message: fmt.Sprintf("on disk but not in index: %s/%s", ref.Library, ref.Topic),
			})
		}
	}
	for _, ref := range onDisk {
		p, err := m.Read(ref.Library, ref.Topic)
		if err != nil {
			continue
		}
		if p.Status == "deprecated" {
			result.Issues = append(result.Issues, LintIssue{
				Kind:    "deprecated",
				Page:    ref,
				Message: fmt.Sprintf("page marked deprecated: %s/%s", ref.Library, ref.Topic),
			})
		}
	}
	return result, nil
}

// AppendLog adds "[KIND] YYYY-MM-DD: message" to log.md.
func (m *Manager) AppendLog(kind, message string) error {
	if err := os.MkdirAll(m.Root, 0755); err != nil {
		return err
	}
	logPath := filepath.Join(m.Root, "log.md")
	entry := fmt.Sprintf("[%s] %s: %s\n", kind, time.Now().Format("2006-01-02"), message)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(entry)
	return err
}

// updateIndex updates or inserts the entry for page in index.md, preserving
// other sections and entries. The placeholder text from init_cmd's setupWiki
// is replaced on first ingest.
func (m *Manager) updateIndex(page Page) error {
	indexPath := filepath.Join(m.Root, "index.md")
	var content string
	if b, err := os.ReadFile(indexPath); err == nil {
		content = string(b)
	}

	sections := parseIndexSections(content)
	sec := sections[page.Library]
	sec.lib = page.Library
	found := false
	for i := range sec.entries {
		if sec.entries[i].topic == page.Topic {
			sec.entries[i].description = page.Description
			found = true
			break
		}
	}
	if !found {
		sec.entries = append(sec.entries, indexEntry{
			topic:       page.Topic,
			description: page.Description,
		})
	}
	sections[page.Library] = sec

	return os.WriteFile(indexPath, []byte(renderIndex(sections)), 0644)
}

type indexEntry struct {
	topic       string
	description string
}

type indexSection struct {
	lib     string
	entries []indexEntry
}

var (
	sectionHeaderRe = regexp.MustCompile(`^## (.+?)\s*$`)
	indexEntryRe    = regexp.MustCompile(`^- \[(.+?)\]\(\./[^/]+/[^)]+\.md\)(?:\s+-\s+(.*))?$`)
)

func parseIndexSections(content string) map[string]indexSection {
	result := map[string]indexSection{}
	var currentLib string
	for line := range strings.SplitSeq(content, "\n") {
		if m := sectionHeaderRe.FindStringSubmatch(line); m != nil {
			currentLib = strings.TrimSpace(m[1])
			if _, ok := result[currentLib]; !ok {
				result[currentLib] = indexSection{lib: currentLib}
			}
			continue
		}
		if currentLib == "" {
			continue
		}
		if m := indexEntryRe.FindStringSubmatch(line); m != nil {
			sec := result[currentLib]
			sec.entries = append(sec.entries, indexEntry{
				topic:       m[1],
				description: strings.TrimSpace(m[2]),
			})
			result[currentLib] = sec
		}
	}
	return result
}

func renderIndex(sections map[string]indexSection) string {
	libs := make([]string, 0, len(sections))
	for lib := range sections {
		libs = append(libs, lib)
	}
	sort.Strings(libs)
	var b strings.Builder
	b.WriteString("# Project Wiki Index\n\n")
	for _, lib := range libs {
		sec := sections[lib]
		if len(sec.entries) == 0 {
			continue
		}
		b.WriteString("## ")
		b.WriteString(lib)
		b.WriteString("\n\n")
		sort.Slice(sec.entries, func(i, j int) bool {
			return sec.entries[i].topic < sec.entries[j].topic
		})
		for _, e := range sec.entries {
			line := fmt.Sprintf("- [%s](./%s/%s.md)", e.topic, lib, e.topic)
			if e.description != "" {
				line += " - " + e.description
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m *Manager) parseIndexEntries() ([]PageRef, error) {
	indexPath := filepath.Join(m.Root, "index.md")
	b, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	sections := parseIndexSections(string(b))
	var refs []PageRef
	for lib, sec := range sections {
		for _, e := range sec.entries {
			refs = append(refs, PageRef{
				Library:     lib,
				Topic:       e.topic,
				Path:        fmt.Sprintf("%s/%s.md", lib, e.topic),
				Description: e.description,
			})
		}
	}
	return refs, nil
}

func renderPage(p Page) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "status: %s\n", p.Status)
	fmt.Fprintf(&b, "updated: %s\n", p.Updated)
	b.WriteString("related: [")
	b.WriteString(strings.Join(p.Related, ", "))
	b.WriteString("]\n")
	b.WriteString("refs: [")
	b.WriteString(strings.Join(p.Refs, ", "))
	b.WriteString("]\n")
	if p.Description != "" {
		fmt.Fprintf(&b, "description: %s\n", p.Description)
	}
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# %s - %s\n\n", p.Library, p.Topic)
	b.WriteString(p.Body)
	if !strings.HasSuffix(p.Body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

func parsePage(content string) Page {
	p := Page{}
	rest, ok := strings.CutPrefix(content, "---\n")
	if !ok {
		p.Body = content
		return p
	}
	fm, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		p.Body = content
		return p
	}
	for line := range strings.SplitSeq(fm, "\n") {
		if v, ok := strings.CutPrefix(line, "status:"); ok {
			p.Status = strings.TrimSpace(v)
		} else if v, ok := strings.CutPrefix(line, "updated:"); ok {
			p.Updated = strings.TrimSpace(v)
		} else if v, ok := strings.CutPrefix(line, "description:"); ok {
			p.Description = strings.TrimSpace(v)
		} else if v, ok := strings.CutPrefix(line, "related:"); ok {
			p.Related = parseListField(v)
		} else if v, ok := strings.CutPrefix(line, "refs:"); ok {
			p.Refs = parseListField(v)
		}
	}
	body = strings.TrimLeft(body, "\n")
	p.Body = body
	return p
}

func parseListField(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func extractFirstDescLine(body string) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) > 120 {
			line = line[:117] + "..."
		}
		return line
	}
	return ""
}

func slug(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.TrimLeft(s, "./")
	return s
}
