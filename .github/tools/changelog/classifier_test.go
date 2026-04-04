package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClassifyFeat(t *testing.T) {
	e := ClassifyCommit("feat(cli): add --format flag for JSON output", "abc123", "coderank", "alice")
	assert.Equal(t, CategoryFeatures, e.Category)
	assert.Equal(t, "cli", e.Scope)
	assert.Equal(t, "add --format flag for JSON output", e.Summary)
	assert.Equal(t, "coderank", e.Repo)
}

func TestClassifyFix(t *testing.T) {
	e := ClassifyCommit("fix: prevent panic on empty library config", "def456", "pipeline", "bob")
	assert.Equal(t, CategoryFixes, e.Category)
	assert.Equal(t, "", e.Scope)
	assert.Equal(t, "prevent panic on empty library config", e.Summary)
}

func TestClassifyLibraryScope(t *testing.T) {
	e := ClassifyCommit("feat(library): add support for htmx@2.0", "ghi789", "pipeline", "alice")
	assert.Equal(t, CategoryLibraries, e.Category)
}

func TestClassifyLibrariesScope(t *testing.T) {
	e := ClassifyCommit("feat(libraries): add 12 new Go libraries", "ghi790", "pipeline", "alice")
	assert.Equal(t, CategoryLibraries, e.Category)
}

func TestClassifyBreakingBang(t *testing.T) {
	e := ClassifyCommit("feat!: rename query command to search", "jkl012", "coderank", "bob")
	assert.Equal(t, CategoryBreaking, e.Category)
}

func TestClassifyBreakingFooter(t *testing.T) {
	e := ClassifyCommit("feat: change API response format\n\nBREAKING CHANGE: response envelope changed", "mno345", "api", "alice")
	assert.Equal(t, CategoryBreaking, e.Category)
}

func TestClassifyNonConventional(t *testing.T) {
	e := ClassifyCommit("Update README", "pqr678", "pipeline", "bob")
	assert.Equal(t, CategoryOther, e.Category)
	assert.Equal(t, "Update README", e.Summary)
}

func TestClassifyPerf(t *testing.T) {
	e := ClassifyCommit("perf(embed): batch Voyage AI calls to reduce latency", "stu901", "pipeline", "alice")
	assert.Equal(t, CategoryPerf, e.Category)
	assert.Equal(t, "embed", e.Scope)
}

func TestClassifyRefactorIsImprovement(t *testing.T) {
	e := ClassifyCommit("refactor(cache): simplify eviction logic", "vwx234", "coderank", "alice")
	assert.Equal(t, CategoryImproved, e.Category)
}

func TestClassifySkipsDocsChoreTest(t *testing.T) {
	for _, msg := range []string{
		"docs: update API reference",
		"chore: bump Go to 1.23",
		"test: add fixture for edge case",
		"ci: fix flaky workflow",
	} {
		e := ClassifyCommit(msg, "aaa", "api", "alice")
		assert.Equal(t, CategoryOther, e.Category, "expected Other for: %s", msg)
	}
}

func TestBuildReleaseFiltersOther(t *testing.T) {
	entries := []Entry{
		{Category: CategoryFeatures, Summary: "add search command"},
		{Category: CategoryOther, Summary: "update README"},
		{Category: CategoryFixes, Summary: "fix timeout"},
	}
	r := BuildRelease("v0.2.0", time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), entries)
	assert.Equal(t, 2, r.Stats.TotalChanges, "Other entries must be filtered")
	assert.Empty(t, r.ByCategory[CategoryOther])
}

func TestBuildReleaseStats(t *testing.T) {
	entries := []Entry{
		{Category: CategoryLibraries, Summary: "add react"},
		{Category: CategoryLibraries, Summary: "add vue"},
		{Category: CategoryFixes, Summary: "fix crash"},
		{Category: CategoryBreaking, Summary: "rename flag"},
	}
	r := BuildRelease("v0.3.0", time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC), entries)
	assert.Equal(t, 4, r.Stats.TotalChanges)
	assert.Equal(t, 2, r.Stats.NewLibraries)
	assert.Equal(t, 1, r.Stats.BugFixes)
	assert.Equal(t, 1, r.Stats.BreakingChanges)
	// Version prefix stripped
	assert.Equal(t, "0.3.0", r.Version)
}

func TestToMarkdown(t *testing.T) {
	entries := []Entry{
		{Category: CategoryFeatures, Scope: "cli", Summary: "add search command"},
		{Category: CategoryFixes, Summary: "fix timeout on large repos"},
		{Category: CategoryLibraries, Summary: "add support for 50 new libraries"},
	}
	r := BuildRelease("v0.2.0", time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), entries)
	md := r.ToMarkdown()
	assert.Contains(t, md, "## v0.2.0 (2025-06-15)")
	assert.Contains(t, md, "### New Libraries")
	assert.Contains(t, md, "### Features")
	assert.Contains(t, md, "**cli:** add search command")
	assert.Contains(t, md, "### Bug Fixes")
	// Breaking before Libraries in output
	assert.Less(t, strings.Index(md, "New Libraries"), strings.Index(md, "Features"))
}

func TestToMDXFrontmatter(t *testing.T) {
	entries := []Entry{
		{Category: CategoryFeatures, Summary: "new thing"},
		{Category: CategoryBreaking, Summary: "API change"},
	}
	r := BuildRelease("v1.0.0", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), entries)
	mdx := r.ToMDX()
	assert.Contains(t, mdx, `version: "1.0.0"`)
	assert.Contains(t, mdx, `date: "2025-07-01"`)
	assert.Contains(t, mdx, `"Breaking Changes"`)
	assert.Contains(t, mdx, `"Features"`)
	assert.Contains(t, mdx, "## Breaking Changes")
	assert.Contains(t, mdx, "## Features")
	// Breaking Changes appear before Features in body
	assert.Less(t, strings.Index(mdx, "Breaking Changes"), strings.Index(mdx, "## Features"))
}
