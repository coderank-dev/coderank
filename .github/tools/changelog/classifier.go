package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Category represents a user-facing changelog category.
type Category string

const (
	CategoryBreaking  Category = "Breaking Changes"
	CategoryLibraries Category = "New Libraries"
	CategoryFeatures  Category = "Features"
	CategoryImproved  Category = "Improvements"
	CategoryPerf      Category = "Performance"
	CategoryFixes     Category = "Bug Fixes"
	CategoryOther     Category = "Other"
)

// CategoryOrder defines display order — breaking first, other last.
var CategoryOrder = []Category{
	CategoryBreaking,
	CategoryLibraries,
	CategoryFeatures,
	CategoryImproved,
	CategoryPerf,
	CategoryFixes,
	CategoryOther,
}

// Entry is a single changelog item parsed from a conventional commit.
type Entry struct {
	Category    Category `json:"category"`
	Scope       string   `json:"scope,omitempty"`
	Summary     string   `json:"summary"`
	CommitHash  string   `json:"commit_hash"`
	Repo        string   `json:"repo"`
	AuthorLogin string   `json:"author_login,omitempty"`
}

// Release is the complete changelog for one product version.
type Release struct {
	Version    string               `json:"version"`
	Date       string               `json:"date"`
	Entries    []Entry              `json:"entries"`
	ByCategory map[Category][]Entry `json:"by_category"`
	Stats      ReleaseStats         `json:"stats"`
}

// ReleaseStats holds summary counters for the release.
type ReleaseStats struct {
	TotalChanges    int `json:"total_changes"`
	NewLibraries    int `json:"new_libraries"`
	BugFixes        int `json:"bug_fixes"`
	BreakingChanges int `json:"breaking_changes"`
}

// conventional commit: type(scope): summary  or  type!: summary
var conventionalRe = regexp.MustCompile(`^(\w+)(?:\(([^)]*)\))?(!)?\s*:\s*(.+)`)

// ClassifyCommit parses a conventional commit message and returns an Entry.
func ClassifyCommit(message, hash, repo, authorLogin string) Entry {
	isBreaking := strings.Contains(message, "BREAKING CHANGE") ||
		strings.Contains(message, "BREAKING-CHANGE")

	firstLine := strings.SplitN(message, "\n", 2)[0]
	m := conventionalRe.FindStringSubmatch(firstLine)

	entry := Entry{CommitHash: hash, Repo: repo, AuthorLogin: authorLogin}

	if m == nil {
		entry.Category = CategoryOther
		entry.Summary = firstLine
		return entry
	}

	commitType := strings.ToLower(m[1])
	scope := m[2]
	bang := m[3]
	summary := m[4]

	entry.Scope = scope
	entry.Summary = summary

	if isBreaking || bang == "!" {
		entry.Category = CategoryBreaking
		return entry
	}

	if scope == "library" || scope == "libraries" || scope == "registry" {
		entry.Category = CategoryLibraries
		return entry
	}

	switch commitType {
	case "feat":
		entry.Category = CategoryFeatures
	case "fix":
		entry.Category = CategoryFixes
	case "perf":
		entry.Category = CategoryPerf
	case "refactor":
		entry.Category = CategoryImproved
	default:
		// docs, test, chore, ci, build — not user-facing
		entry.Category = CategoryOther
	}
	return entry
}

// BuildRelease groups entries by category, filters non-user-facing, computes stats.
func BuildRelease(version string, date time.Time, entries []Entry) *Release {
	byCategory := make(map[Category][]Entry)
	var userFacing []Entry

	for _, e := range entries {
		if e.Category == CategoryOther {
			continue
		}
		userFacing = append(userFacing, e)
		byCategory[e.Category] = append(byCategory[e.Category], e)
	}

	// Sort entries within each category for deterministic output.
	for cat := range byCategory {
		sort.Slice(byCategory[cat], func(i, j int) bool {
			return byCategory[cat][i].Summary < byCategory[cat][j].Summary
		})
	}

	return &Release{
		Version:    strings.TrimPrefix(version, "v"),
		Date:       date.Format("2006-01-02"),
		Entries:    userFacing,
		ByCategory: byCategory,
		Stats: ReleaseStats{
			TotalChanges:    len(userFacing),
			NewLibraries:    len(byCategory[CategoryLibraries]),
			BugFixes:        len(byCategory[CategoryFixes]),
			BreakingChanges: len(byCategory[CategoryBreaking]),
		},
	}
}

// ToMarkdown renders the release as a markdown changelog entry (GitHub Releases).
func (r *Release) ToMarkdown() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## v%s (%s)\n\n", r.Version, r.Date))

	for _, cat := range CategoryOrder {
		entries := r.ByCategory[cat]
		if len(entries) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s\n\n", cat))
		for _, e := range entries {
			if e.Scope != "" {
				b.WriteString(fmt.Sprintf("- **%s:** %s\n", e.Scope, e.Summary))
			} else {
				b.WriteString(fmt.Sprintf("- %s\n", e.Summary))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ToMDX renders the release as an MDX file with frontmatter for the web changelog page.
func (r *Release) ToMDX() string {
	var b strings.Builder

	var tags []string
	for _, cat := range CategoryOrder {
		if len(r.ByCategory[cat]) > 0 {
			tags = append(tags, fmt.Sprintf("%q", string(cat)))
		}
	}

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("version: %q\n", r.Version))
	b.WriteString(fmt.Sprintf("date: %q\n", r.Date))
	b.WriteString(fmt.Sprintf("totalChanges: %d\n", r.Stats.TotalChanges))
	b.WriteString(fmt.Sprintf("newLibraries: %d\n", r.Stats.NewLibraries))
	b.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(tags, ", ")))
	b.WriteString("---\n\n")

	for _, cat := range CategoryOrder {
		entries := r.ByCategory[cat]
		if len(entries) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("## %s\n\n", cat))
		for _, e := range entries {
			if e.Scope != "" {
				b.WriteString(fmt.Sprintf("- **%s:** %s\n", e.Scope, e.Summary))
			} else {
				b.WriteString(fmt.Sprintf("- %s\n", e.Summary))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}
