package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestInitGeneratesValidYAML(t *testing.T) {
	config := CodeRankConfig{
		Stack: StackConfig{
			Language:  "typescript",
			Framework: "nextjs",
			Preferred: []string{"prisma", "zod"},
			Blocked:   []string{"moment"},
		},
		Context: ContextConfig{
			MaxTokens:        5000,
			IncludeMigration: true,
			IncludeExamples:  true,
			PreferTypeScript: true,
		},
		Curation: CurationConfig{
			MinHealthScore:   60,
			PreferMaintained: true,
			SecurityBlock:    "critical",
		},
		Inject: InjectConfig{
			Target: "auto",
			Agent:  "auto",
		},
	}

	data, err := yaml.Marshal(&config)
	require.NoError(t, err)

	var roundTrip CodeRankConfig
	require.NoError(t, yaml.Unmarshal(data, &roundTrip))
	assert.Equal(t, "typescript", roundTrip.Stack.Language,
		"language should round-trip through YAML")
	assert.Equal(t, []string{"prisma", "zod"}, roundTrip.Stack.Preferred,
		"preferred libs should round-trip through YAML")
	assert.Equal(t, "critical", roundTrip.Curation.SecurityBlock,
		"all sections should be present in output YAML")
}

func TestDetectLanguageFromPackageJSON(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	require.NoError(t, os.WriteFile("package.json", []byte("{}"), 0644))
	assert.Equal(t, "typescript", detectLanguage(),
		"should detect TypeScript from package.json")
}

func TestDetectLanguageFromGoMod(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	require.NoError(t, os.WriteFile("go.mod", []byte("module example.com/app\n\ngo 1.22\n"), 0644))
	assert.Equal(t, "go", detectLanguage(),
		"should detect Go from go.mod")
}

func TestDetectLanguageReturnsEmptyWhenUnknown(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	assert.Equal(t, "", detectLanguage(),
		"should return empty string when no known project files found")
}

func TestSplitAndTrim(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"prisma, zod , tailwindcss", []string{"prisma", "zod", "tailwindcss"}},
		{"moment", []string{"moment"}},
		{"  a , , b  ", []string{"a", "b"}}, // empty entries dropped
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, splitAndTrim(tc.input),
			"splitAndTrim(%q) should trim spaces and drop empty entries", tc.input)
	}
}
