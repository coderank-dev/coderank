package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/coderank-dev/coderank/internal/agents"
	"github.com/coderank-dev/coderank/internal/render"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// CodeRankConfig represents the .coderank.yml file structure.
type CodeRankConfig struct {
	Stack    StackConfig    `yaml:"stack"`
	Context  ContextConfig  `yaml:"context"`
	Curation CurationConfig `yaml:"curation"`
	Inject   InjectConfig   `yaml:"inject"`
	Wiki     WikiConfig     `yaml:"wiki"`
}

// StackConfig defines the project's technology stack.
type StackConfig struct {
	Language  string   `yaml:"language"`
	Runtime   string   `yaml:"runtime,omitempty"`
	Framework string   `yaml:"framework,omitempty"`
	Preferred []string `yaml:"preferred,omitempty"`
	Blocked   []string `yaml:"blocked,omitempty"`
}

// ContextConfig defines query behavior defaults.
type ContextConfig struct {
	MaxTokens        int  `yaml:"max_tokens"`
	IncludeMigration bool `yaml:"include_migration"`
	IncludeExamples  bool `yaml:"include_examples"`
	PreferTypeScript bool `yaml:"prefer_typescript"`
}

// CurationConfig defines library quality thresholds.
type CurationConfig struct {
	MinHealthScore   int    `yaml:"min_health_score"`
	PreferMaintained bool   `yaml:"prefer_maintained"`
	SecurityBlock    string `yaml:"security_block"`
}

// InjectConfig defines ambient inject behavior.
type InjectConfig struct {
	Target string `yaml:"target"`
	Watch  bool   `yaml:"watch"`
	Agent  string `yaml:"agent"`
}

// WikiConfig controls the project wiki behavior.
type WikiConfig struct {
	Enabled bool `yaml:"enabled"`
}

// initCmd generates a .coderank.yml config file via interactive wizard,
// creates the .coderank/wiki/ directory, and installs agent skills.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize CodeRank in the current project",
	Long: `Interactive wizard to set up CodeRank in the current project.

Creates:
  .coderank.yml          — project config
  .coderank/wiki/        — project knowledge wiki
  <agent>/skills/coderank/       — root CodeRank skill
  <agent>/skills/coderank-wiki/  — wiki skill

Use --non-interactive with flags for CI/script usage.
Use --no-wiki to skip wiki setup.

Examples:
  coderank init
  coderank init --non-interactive --language typescript --framework nextjs
  coderank init --no-wiki`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().Bool("non-interactive", false, "Skip interactive wizard, use flags")
	initCmd.Flags().String("language", "", "Programming language (typescript, go, python, rust)")
	initCmd.Flags().String("framework", "", "Framework (nextjs, react, gin, fastapi, etc.)")
	initCmd.Flags().StringSlice("preferred", nil, "Preferred libraries")
	initCmd.Flags().StringSlice("blocked", nil, "Blocked libraries")
	initCmd.Flags().Int("max-tokens", 500, "Default token budget per query")
	initCmd.Flags().Bool("no-wiki", false, "Skip wiki setup")
}

func runInit(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	noWiki, _ := cmd.Flags().GetBool("no-wiki")

	// Check for existing config
	if _, err := os.Stat(".coderank.yml"); err == nil {
		if !nonInteractive {
			var overwrite bool
			huh.NewConfirm().
				Title(".coderank.yml already exists. Overwrite?").
				Value(&overwrite).
				Run()
			if !overwrite {
				fmt.Print(render.WarningMsg("Init cancelled — existing config preserved"))
				return nil
			}
		}
	}

	config := CodeRankConfig{
		Context: ContextConfig{
			MaxTokens:        5000,
			IncludeMigration: true,
			IncludeExamples:  true,
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
		Wiki: WikiConfig{
			Enabled: !noWiki,
		},
	}

	if nonInteractive {
		config.Stack.Language, _ = cmd.Flags().GetString("language")
		config.Stack.Framework, _ = cmd.Flags().GetString("framework")
		config.Stack.Preferred, _ = cmd.Flags().GetStringSlice("preferred")
		config.Stack.Blocked, _ = cmd.Flags().GetStringSlice("blocked")
		config.Context.MaxTokens, _ = cmd.Flags().GetInt("max-tokens")
	} else {
		detectedLang := detectLanguage()

		var language, framework, preferredStr, blockedStr string
		if detectedLang != "" {
			language = detectedLang
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What language does this project use?").
					Options(
						huh.NewOption("TypeScript", "typescript"),
						huh.NewOption("Go", "go"),
						huh.NewOption("Python", "python"),
						huh.NewOption("Rust", "rust"),
						huh.NewOption("Other", "other"),
					).
					Value(&language),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("Framework (optional)").
					Placeholder("e.g., nextjs, react, gin, fastapi").
					Value(&framework),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("Preferred libraries (comma-separated)").
					Placeholder("e.g., prisma, zod, tailwindcss").
					Value(&preferredStr),
				huh.NewInput().
					Title("Blocked libraries (comma-separated)").
					Placeholder("e.g., moment, enzyme, lodash").
					Value(&blockedStr),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		config.Stack.Language = language
		config.Stack.Framework = framework
		if preferredStr != "" {
			config.Stack.Preferred = splitAndTrim(preferredStr)
		}
		if blockedStr != "" {
			config.Stack.Blocked = splitAndTrim(blockedStr)
		}
		if language == "typescript" {
			config.Context.PreferTypeScript = true
		}
	}

	// Write .coderank.yml
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(".coderank.yml", data, 0644); err != nil {
		return fmt.Errorf("writing .coderank.yml: %w", err)
	}
	fmt.Print(render.SuccessMsg("Created .coderank.yml"))

	// Create wiki directory
	if !noWiki {
		if err := setupWiki(); err != nil {
			return fmt.Errorf("setting up wiki: %w", err)
		}
	}

	// Install skills into detected agents
	if err := installSkills(noWiki); err != nil {
		return fmt.Errorf("installing skills: %w", err)
	}

	return nil
}

// setupWiki creates .coderank/wiki/ with index.md and log.md.
func setupWiki() error {
	wikiDir := filepath.Join(".coderank", "wiki")
	if err := os.MkdirAll(wikiDir, 0755); err != nil {
		return fmt.Errorf("creating wiki directory: %w", err)
	}

	indexContent := "# Project Wiki Index\n\nPages will appear here as you use `coderank query`.\n"
	if err := writeIfNotExists(filepath.Join(wikiDir, "index.md"), indexContent); err != nil {
		return fmt.Errorf("writing index.md: %w", err)
	}

	logContent := fmt.Sprintf("# Wiki Log\n\n[INIT] %s: wiki initialized\n", time.Now().Format("2006-01-02"))
	if err := writeIfNotExists(filepath.Join(wikiDir, "log.md"), logContent); err != nil {
		return fmt.Errorf("writing log.md: %w", err)
	}

	fmt.Print(render.SuccessMsg("Created .coderank/wiki/"))
	return nil
}

// installSkills installs the root skill and optionally the wiki skill
// into all detected AI agents in the current project.
func installSkills(noWiki bool) error {
	projectRoot, _ := os.Getwd()
	detectedAgents := agents.Detect(projectRoot)

	if len(detectedAgents) == 0 {
		fmt.Print(render.WarningMsg("No AI agents detected — skipping skill installation"))
		fmt.Fprintln(os.Stderr, "  Run 'coderank install <lib>' manually after setting up an agent.")
		return nil
	}

	agentNames := make([]string, len(detectedAgents))
	for i, a := range detectedAgents {
		agentNames[i] = a.Name
	}
	fmt.Fprintf(os.Stderr, "Installing skills → %s\n", strings.Join(agentNames, ", "))

	rootContent := agents.RootSkillMD()
	wikiContent := agents.WikiSkillMD()

	for _, agent := range detectedAgents {
		rootPath := agents.SkillPath(projectRoot, agent, agents.RootSkillName, false)
		if err := agents.WriteSkill(rootPath, rootContent); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s/%s: %v\n", agent.ID, agents.RootSkillName, err)
		}

		if !noWiki {
			wikiPath := agents.SkillPath(projectRoot, agent, agents.WikiSkillName, false)
			if err := agents.WriteSkill(wikiPath, wikiContent); err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s/%s: %v\n", agent.ID, agents.WikiSkillName, err)
			}
		}
	}

	fmt.Print(render.SuccessMsg("Installed CodeRank skills"))
	return nil
}

// writeIfNotExists writes content to path only if the file does not already exist,
// preserving any existing content on re-runs of coderank init.
func writeIfNotExists(path, content string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// detectLanguage checks for common project files to auto-detect the language.
func detectLanguage() string {
	if _, err := os.Stat("package.json"); err == nil {
		return "typescript"
	}
	if _, err := os.Stat("go.mod"); err == nil {
		return "go"
	}
	if _, err := os.Stat("requirements.txt"); err == nil {
		return "python"
	}
	if _, err := os.Stat("pyproject.toml"); err == nil {
		return "python"
	}
	if _, err := os.Stat("Cargo.toml"); err == nil {
		return "rust"
	}
	return ""
}

// splitAndTrim splits a comma-separated string and trims whitespace.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
