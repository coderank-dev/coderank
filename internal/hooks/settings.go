// Package hooks manages the hooks section of Claude Code's .claude/settings.json.
// It provides idempotent add/remove of coderank-managed hook entries while
// preserving any other settings the user or other tools have written.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Event names from Claude Code's hooks schema that coderank targets.
const (
	EventUserPromptSubmit = "UserPromptSubmit"
	EventPostToolUse      = "PostToolUse"
)

// coderankMarker is the substring we detect on existing command strings to
// know a hook entry is coderank-owned and can be updated or removed.
const coderankMarker = "coderank wiki hook"

// HookEntry is a single hook handler, matching Claude Code's command-hook shape.
type HookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// MatcherGroup groups hook entries that share a matcher.
type MatcherGroup struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []HookEntry `json:"hooks"`
}

// Settings is a partial view of .claude/settings.json focused on the hooks key.
// Other top-level keys are preserved verbatim across load/save via Passthrough.
type Settings struct {
	Hooks       map[string][]MatcherGroup
	Passthrough map[string]json.RawMessage
}

// Load reads settings.json from path. Returns an empty Settings if the file
// doesn't exist - callers can then add hooks and Save to create it.
func Load(path string) (*Settings, error) {
	s := &Settings{
		Hooks:       map[string][]MatcherGroup{},
		Passthrough: map[string]json.RawMessage{},
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	for k, v := range raw {
		if k == "hooks" {
			if err := json.Unmarshal(v, &s.Hooks); err != nil {
				return nil, fmt.Errorf("parsing hooks section: %w", err)
			}
			continue
		}
		s.Passthrough[k] = v
	}
	if s.Hooks == nil {
		s.Hooks = map[string][]MatcherGroup{}
	}
	return s, nil
}

// Save writes the settings to path, preserving Passthrough fields and creating
// parent directories as needed.
func (s *Settings) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out := map[string]any{}
	for k, v := range s.Passthrough {
		out[k] = v
	}
	if len(s.Hooks) > 0 {
		out["hooks"] = s.Hooks
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

// AddCoderankHook idempotently registers a coderank-owned hook entry for the
// given event and matcher. If a coderank entry already exists on the same
// matcher group, its command is updated in place so re-runs of coderank init
// don't duplicate hooks. Non-coderank hook entries in the file are preserved.
func (s *Settings) AddCoderankHook(event, matcher, command string) {
	if s.Hooks == nil {
		s.Hooks = map[string][]MatcherGroup{}
	}
	groups := s.Hooks[event]
	for gi, group := range groups {
		if group.Matcher == matcher {
			for hi, h := range group.Hooks {
				if isCoderankCommand(h.Command) {
					groups[gi].Hooks[hi].Command = command
					s.Hooks[event] = groups
					return
				}
			}
			groups[gi].Hooks = append(groups[gi].Hooks, HookEntry{Type: "command", Command: command})
			s.Hooks[event] = groups
			return
		}
	}
	s.Hooks[event] = append(groups, MatcherGroup{
		Matcher: matcher,
		Hooks:   []HookEntry{{Type: "command", Command: command}},
	})
}

// RemoveCoderankHooks deletes every coderank-owned hook entry, leaving user
// and third-party entries untouched. Empty matcher groups are pruned, and
// event keys with no remaining groups are removed entirely.
func (s *Settings) RemoveCoderankHooks() {
	for event, groups := range s.Hooks {
		newGroups := make([]MatcherGroup, 0, len(groups))
		for _, group := range groups {
			kept := make([]HookEntry, 0, len(group.Hooks))
			for _, h := range group.Hooks {
				if !isCoderankCommand(h.Command) {
					kept = append(kept, h)
				}
			}
			if len(kept) > 0 {
				group.Hooks = kept
				newGroups = append(newGroups, group)
			}
		}
		if len(newGroups) > 0 {
			s.Hooks[event] = newGroups
		} else {
			delete(s.Hooks, event)
		}
	}
}

func isCoderankCommand(cmd string) bool {
	return strings.Contains(cmd, coderankMarker)
}
