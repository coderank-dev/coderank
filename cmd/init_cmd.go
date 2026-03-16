package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
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

// initCmd generates a .coderank.yml config file via interactive wizard.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a .coderank.yml config file",
	Long: `Interactive wizard to generate a .coderank.yml config file for your project.
Detects your language and framework automatically from package.json or go.mod.

Use --non-interactive with flags for CI/script usage.

Examples:
  coderank init
  coderank init --non-interactive --language typescript --framework nextjs`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().Bool("non-interactive", false, "Skip interactive wizard, use flags")
	initCmd.Flags().String("language", "", "Programming language (typescript, go, python, rust)")
	initCmd.Flags().String("framework", "", "Framework (nextjs, react, gin, fastapi, etc.)")
	initCmd.Flags().StringSlice("preferred", nil, "Preferred libraries")
	initCmd.Flags().StringSlice("blocked", nil, "Blocked libraries")
	initCmd.Flags().Int("max-tokens", 5000, "Default token budget per query")
}

func runInit(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

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

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(".coderank.yml", data, 0644); err != nil {
		return fmt.Errorf("writing .coderank.yml: %w", err)
	}

	fmt.Print(render.SuccessMsg("Created .coderank.yml"))
	return nil
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
