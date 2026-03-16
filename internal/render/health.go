package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/coderank-dev/coderank/internal/api"
)

// HealthDisplay renders a styled health score breakdown. Scores are
// color-coded: green (≥70), yellow (40–69), red (<40).
func HealthDisplay(h *api.HealthResponse) string {
	var sb strings.Builder

	// Title with overall score
	sb.WriteString(Title.Render(fmt.Sprintf("%s — %d/100", h.Library, h.HealthScore)))
	sb.WriteString(" ")
	sb.WriteString(scoreLabel(h.HealthScore))
	sb.WriteString("\n\n")

	// Sub-scores with colored bars
	categories := []struct {
		name string
		key  string
	}{
		{"Maintenance", "maintenance"},
		{"Security", "security"},
		{"Community", "community"},
		{"Sustainability", "sustainability"},
	}

	for _, cat := range categories {
		score, ok := h.Breakdown[cat.key]
		if !ok {
			continue
		}
		bar := scoreBar(score)
		sb.WriteString(fmt.Sprintf("  %-16s %s %s\n",
			cat.name, bar, scoreStyle(score).Render(fmt.Sprintf("%d", score))))
	}

	if h.Repo != "" {
		sb.WriteString("\n")
		sb.WriteString(Subtle.Render(fmt.Sprintf("  repo: %s", h.Repo)))
		sb.WriteString("\n")
	}

	if h.LastIndexed != "" {
		sb.WriteString(Subtle.Render(fmt.Sprintf("  last indexed: %s", h.LastIndexed)))
		sb.WriteString("\n")
	}

	return sb.String()
}

// scoreBar renders a visual bar proportional to the score (0–100).
// Uses 20 characters: filled blocks (█) + empty blocks (░).
func scoreBar(score int) string {
	filled := score / 5 // 0–20 characters
	empty := 20 - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return scoreStyle(score).Render(bar)
}

// scoreLabel returns a styled severity label.
func scoreLabel(score int) string {
	if score >= 70 {
		return Success.Render("HEALTHY")
	}
	if score >= 40 {
		return Warning.Render("CAUTION")
	}
	return Error.Render("LOW")
}

// scoreStyle returns the Lip Gloss style for a score value.
func scoreStyle(score int) lipgloss.Style {
	if score >= 70 {
		return Success
	}
	if score >= 40 {
		return Warning
	}
	return Error
}
