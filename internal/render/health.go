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

	// Title line: 🏥 react  98/100  ● HEALTHY
	score := scoreStyle(h.HealthScore).Bold(true).Render(fmt.Sprintf("%d/100", h.HealthScore))
	sb.WriteString(fmt.Sprintf("🏥 %s  %s  %s\n\n",
		Title.Render(h.Library), score, scoreLabel(h.HealthScore)))

	// Sub-scores with colored bars
	categories := []struct {
		emoji string
		name  string
		key   string
	}{
		{"🔧", "Maintenance", "maintenance"},
		{"🔒", "Security", "security"},
		{"👥", "Community", "community"},
		{"♻️ ", "Sustainability", "sustainability"},
	}

	for _, cat := range categories {
		score, ok := h.Breakdown[cat.key]
		if !ok {
			continue
		}
		bar := scoreBar(score)
		sb.WriteString(fmt.Sprintf("  %s %-14s %s %s\n",
			cat.emoji, cat.name, bar, scoreStyle(score).Render(fmt.Sprintf("%d", score))))
	}

	if h.Repo != "" {
		sb.WriteString("\n")
		sb.WriteString(Subtle.Render(fmt.Sprintf("  repo: github.com/%s", h.Repo)))
		sb.WriteString("\n")
	}

	if h.LastIndexed != "" {
		sb.WriteString(Subtle.Render(fmt.Sprintf("  indexed: %s", h.LastIndexed)))
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
