package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIngestCreatesPageIndexAndLog(t *testing.T) {
	root := t.TempDir()
	m := New(root)

	page, err := m.Ingest("zod", "auth-schemas", "Discriminated unions for login/signup.", IngestOpts{
		Refs:        []string{"src/auth.ts"},
		Description: "Auth schema patterns",
	})
	require.NoError(t, err)
	assert.Equal(t, "zod", page.Library)
	assert.Equal(t, "auth-schemas", page.Topic)

	pageBytes, err := os.ReadFile(filepath.Join(root, WikiDir, "zod", "auth-schemas.md"))
	require.NoError(t, err)
	content := string(pageBytes)
	assert.Contains(t, content, "status: current")
	assert.Contains(t, content, "refs: [src/auth.ts]")
	assert.Contains(t, content, "description: Auth schema patterns")
	assert.Contains(t, content, "# zod - auth-schemas")
	assert.Contains(t, content, "Discriminated unions")

	idxBytes, err := os.ReadFile(filepath.Join(root, WikiDir, "index.md"))
	require.NoError(t, err)
	idx := string(idxBytes)
	assert.Contains(t, idx, "## zod")
	assert.Contains(t, idx, "[auth-schemas](./zod/auth-schemas.md)")
	assert.Contains(t, idx, "Auth schema patterns")

	logBytes, err := os.ReadFile(filepath.Join(root, WikiDir, "log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(logBytes), "[INGEST]")
	assert.Contains(t, string(logBytes), "zod - auth-schemas")
}

func TestIngestReplacesExistingPageWithoutDuplicatingIndex(t *testing.T) {
	root := t.TempDir()
	m := New(root)

	_, err := m.Ingest("zod", "auth", "first body", IngestOpts{Description: "first"})
	require.NoError(t, err)
	_, err = m.Ingest("zod", "auth", "second body", IngestOpts{Description: "second"})
	require.NoError(t, err)

	pageBytes, err := os.ReadFile(filepath.Join(root, WikiDir, "zod", "auth.md"))
	require.NoError(t, err)
	assert.Contains(t, string(pageBytes), "second body")
	assert.NotContains(t, string(pageBytes), "first body", "old body must be fully replaced")

	idxBytes, err := os.ReadFile(filepath.Join(root, WikiDir, "index.md"))
	require.NoError(t, err)
	idx := string(idxBytes)
	assert.Equal(t, 1, strings.Count(idx, "[auth](./zod/auth.md)"), "index must not duplicate entry")
	assert.Contains(t, idx, "second", "index description must reflect latest ingest")

	logBytes, err := os.ReadFile(filepath.Join(root, WikiDir, "log.md"))
	require.NoError(t, err)
	assert.Equal(t, 2, strings.Count(string(logBytes), "[INGEST]"),
		"log is append-only: two ingests produce two entries")
}

func TestIngestAcrossLibrariesCreatesDistinctSections(t *testing.T) {
	root := t.TempDir()
	m := New(root)
	_, err := m.Ingest("zod", "auth", "body", IngestOpts{})
	require.NoError(t, err)
	_, err = m.Ingest("prisma", "migrations", "body", IngestOpts{})
	require.NoError(t, err)

	idxBytes, err := os.ReadFile(filepath.Join(root, WikiDir, "index.md"))
	require.NoError(t, err)
	idx := string(idxBytes)
	assert.Contains(t, idx, "## prisma", "prisma section must exist")
	assert.Contains(t, idx, "## zod", "zod section must exist")
	// Sections are sorted alphabetically; prisma should come before zod
	assert.Less(t, strings.Index(idx, "## prisma"), strings.Index(idx, "## zod"))
}

func TestIngestStripsInitPlaceholder(t *testing.T) {
	root := t.TempDir()
	m := New(root)
	require.NoError(t, os.MkdirAll(m.Root, 0755))
	placeholder := "# Project Wiki Index\n\nPages will appear here as you use `coderank query`.\n"
	require.NoError(t, os.WriteFile(filepath.Join(m.Root, "index.md"), []byte(placeholder), 0644))

	_, err := m.Ingest("zod", "auth", "body", IngestOpts{})
	require.NoError(t, err)

	idxBytes, err := os.ReadFile(filepath.Join(m.Root, "index.md"))
	require.NoError(t, err)
	idx := string(idxBytes)
	assert.NotContains(t, idx, "Pages will appear here", "placeholder must be replaced on first ingest")
	assert.Contains(t, idx, "[auth](./zod/auth.md)")
}

func TestReadReturnsPage(t *testing.T) {
	root := t.TempDir()
	m := New(root)
	_, err := m.Ingest("react", "hooks", "useCallback vs useMemo", IngestOpts{
		Refs:    []string{"src/App.tsx"},
		Related: []string{"react/effects"},
	})
	require.NoError(t, err)

	page, err := m.Read("react", "hooks")
	require.NoError(t, err)
	assert.Equal(t, "react", page.Library)
	assert.Equal(t, "hooks", page.Topic)
	assert.Equal(t, "current", page.Status)
	assert.Equal(t, []string{"src/App.tsx"}, page.Refs)
	assert.Equal(t, []string{"react/effects"}, page.Related)
	assert.Contains(t, page.Body, "useCallback vs useMemo")
}

func TestReadReturnsHelpfulErrorWhenMissing(t *testing.T) {
	m := New(t.TempDir())
	_, err := m.Read("unknown", "nothing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no page for unknown/nothing")
	assert.Contains(t, err.Error(), "coderank wiki list")
}

func TestListReturnsAllPagesSorted(t *testing.T) {
	root := t.TempDir()
	m := New(root)
	_, err := m.Ingest("zod", "auth", "body", IngestOpts{Description: "z"})
	require.NoError(t, err)
	_, err = m.Ingest("prisma", "migrations", "body", IngestOpts{Description: "p"})
	require.NoError(t, err)
	_, err = m.Ingest("prisma", "schema", "body", IngestOpts{Description: "s"})
	require.NoError(t, err)

	pages, err := m.List()
	require.NoError(t, err)
	require.Len(t, pages, 3)
	// Sorted: prisma/migrations, prisma/schema, zod/auth
	assert.Equal(t, "prisma", pages[0].Library)
	assert.Equal(t, "migrations", pages[0].Topic)
	assert.Equal(t, "prisma", pages[1].Library)
	assert.Equal(t, "schema", pages[1].Topic)
	assert.Equal(t, "zod", pages[2].Library)
}

func TestLogWithTail(t *testing.T) {
	root := t.TempDir()
	m := New(root)
	require.NoError(t, m.AppendLog("INIT", "first"))
	require.NoError(t, m.AppendLog("INGEST", "second"))
	require.NoError(t, m.AppendLog("INGEST", "third"))
	require.NoError(t, m.AppendLog("LINT", "fourth"))

	all, err := m.Log(0)
	require.NoError(t, err)
	assert.Equal(t, 4, strings.Count(all, "["))

	tail2, err := m.Log(2)
	require.NoError(t, err)
	assert.NotContains(t, tail2, "first")
	assert.NotContains(t, tail2, "second")
	assert.Contains(t, tail2, "third")
	assert.Contains(t, tail2, "fourth")
}

func TestLintFlagsMissingOrphanedAndDeprecated(t *testing.T) {
	root := t.TempDir()
	m := New(root)

	_, err := m.Ingest("zod", "auth", "body", IngestOpts{})
	require.NoError(t, err)
	_, err = m.Ingest("prisma", "migrations", "body", IngestOpts{Status: "deprecated"})
	require.NoError(t, err)

	orphanPath := filepath.Join(m.Root, "react", "hooks.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(orphanPath), 0755))
	require.NoError(t, os.WriteFile(orphanPath, []byte("---\nstatus: current\n---\n# react - hooks\n"), 0644))

	idxPath := filepath.Join(m.Root, "index.md")
	idxBytes, err := os.ReadFile(idxPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(idxPath, []byte(string(idxBytes)+"\n## express\n\n- [routing](./express/routing.md)\n"), 0644))

	result, err := m.Lint()
	require.NoError(t, err)

	kinds := map[string]int{}
	for _, issue := range result.Issues {
		kinds[issue.Kind]++
	}
	assert.GreaterOrEqual(t, kinds["missing"], 1, "express/routing is indexed but not on disk")
	assert.GreaterOrEqual(t, kinds["orphaned"], 1, "react/hooks is on disk but not indexed")
	assert.GreaterOrEqual(t, kinds["deprecated"], 1, "prisma/migrations is marked deprecated")
}

func TestIngestAutoCreatesWikiRootIfMissing(t *testing.T) {
	root := t.TempDir()
	m := New(root)
	// No init - .coderank/wiki/ doesn't exist yet.
	_, err := m.Ingest("zod", "auth", "body", IngestOpts{})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(root, WikiDir, "zod", "auth.md"))
	require.NoError(t, err, "wiki root must be created implicitly on first ingest")
}

func TestSlugNormalizesLibAndTopic(t *testing.T) {
	assert.Equal(t, "react", slug("React"))
	assert.Equal(t, "server-side-rendering", slug("Server Side Rendering"))
	assert.Equal(t, "next.js", slug(".Next.JS"))
}

func TestMatchLibraries(t *testing.T) {
	libs := []string{"react", "zod", "prisma"}
	assert.Equal(t, []string{"react"}, MatchLibraries("how does useCallback work in react", libs))
	assert.ElementsMatch(t, []string{"zod", "prisma"}, MatchLibraries("add Zod validation for Prisma schema", libs))
	assert.Empty(t, MatchLibraries("git rebase --interactive syntax", libs))
}
