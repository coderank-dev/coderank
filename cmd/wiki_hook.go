package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coderank-dev/coderank/internal/wiki"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var wikiHookCmd = &cobra.Command{
	Use:    "hook <kind>",
	Short:  "Claude Code hook emitter (internal; not meant for direct use)",
	Hidden: true,
	Long: `Reads a Claude Code hook JSON payload from stdin and prints a
JSON response on stdout that injects coderank reminders into Claude's
context when the payload indicates library-related activity.

Supported kinds:
  user-prompt   UserPromptSubmit hook
  post-edit     PostToolUse hook on Edit|Write|MultiEdit

Intended to be wired into .claude/settings.json by 'coderank init'; users
should not invoke this directly. Exits 0 even on internal errors so a hook
misfire never blocks Claude Code.`,
	Args: cobra.ExactArgs(1),
	RunE: runWikiHook,
}

func init() {
	wikiCmd.AddCommand(wikiHookCmd)
}

// hookResponse is the subset of Claude Code's hook output JSON we emit.
type hookResponse struct {
	HookSpecificOutput *hookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func runWikiHook(cmd *cobra.Command, args []string) error {
	kind := args[0]
	// Defer body to helper so we can trap any error and still exit 0.
	msg := buildHookMessage(kind)
	if msg != "" {
		var event string
		switch kind {
		case "user-prompt":
			event = "UserPromptSubmit"
		case "post-edit":
			event = "PostToolUse"
		default:
			return nil
		}
		resp := hookResponse{
			HookSpecificOutput: &hookSpecificOutput{
				HookEventName:     event,
				AdditionalContext: msg,
			},
		}
		b, err := json.Marshal(resp)
		if err == nil {
			_, _ = os.Stdout.Write(b)
			_, _ = os.Stdout.Write([]byte("\n"))
		}
	}
	return nil
}

func buildHookMessage(kind string) string {
	payload, err := readHookPayload(os.Stdin)
	if err != nil {
		return ""
	}
	libs := preferredLibraries()
	if len(libs) == 0 {
		return ""
	}
	switch kind {
	case "user-prompt":
		matched := wiki.MatchLibraries(payload.Prompt, libs)
		if len(matched) == 0 {
			return ""
		}
		return fmt.Sprintf(
			"Library context: this prompt references %s. Before answering, consult `.coderank/wiki/` (run `coderank wiki list`) for project-specific patterns, and run `coderank query <lib> \"<question>\"` for current library docs. Do not answer library questions from training data.",
			joinQuoted(matched),
		)
	case "post-edit":
		if !isEditingTool(payload.ToolName) {
			return ""
		}
		filePath, content := extractEditedFile(payload.ToolInput)
		if filePath == "" {
			return ""
		}
		matched := wiki.MatchLibraries(content, libs)
		if len(matched) == 0 {
			matched = wiki.MatchLibraries(filePath, libs)
		}
		if len(matched) == 0 {
			return ""
		}
		return fmt.Sprintf(
			"You modified `%s` which uses %s. If this introduces a new pattern, decision, or gotcha, record it with: `coderank wiki ingest --lib <lib> --topic <short-topic> --summary \"...\" --refs %s`. Keep the page short - a one-sentence summary and file pointers, not a library tutorial.",
			filePath, joinQuoted(matched), filePath,
		)
	}
	return ""
}

type hookPayload struct {
	Prompt    string                 `json:"prompt,omitempty"`
	ToolName  string                 `json:"tool_name,omitempty"`
	ToolInput map[string]any `json:"tool_input,omitempty"`
}

func readHookPayload(r io.Reader) (*hookPayload, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return &hookPayload{}, nil
	}
	var p hookPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// preferredLibraries reads stack.preferred from .coderank.yml (via viper).
func preferredLibraries() []string {
	return viper.GetStringSlice("stack.preferred")
}

func isEditingTool(name string) bool {
	switch name {
	case "Edit", "Write", "MultiEdit":
		return true
	}
	return false
}

// extractEditedFile pulls the file path and content from a tool_input map.
// Edit/Write/MultiEdit expose a "file_path" field; Write adds "content";
// Edit/MultiEdit carry "new_string"(s) which suffice for substring matching.
func extractEditedFile(in map[string]any) (string, string) {
	if in == nil {
		return "", ""
	}
	path, _ := in["file_path"].(string)
	if c, ok := in["content"].(string); ok {
		return path, c
	}
	if s, ok := in["new_string"].(string); ok {
		return path, s
	}
	if edits, ok := in["edits"].([]any); ok {
		var merged strings.Builder
		for _, e := range edits {
			if em, ok := e.(map[string]any); ok {
				if s, ok := em["new_string"].(string); ok {
					merged.WriteString(s)
					merged.WriteByte('\n')
				}
			}
		}
		return path, merged.String()
	}
	return path, ""
}

func joinQuoted(libs []string) string {
	var b strings.Builder
	for i, l := range libs {
		if i > 0 {
			if i == len(libs)-1 {
				b.WriteString(" and ")
			} else {
				b.WriteString(", ")
			}
		}
		b.WriteByte('`')
		b.WriteString(l)
		b.WriteByte('`')
	}
	return b.String()
}
