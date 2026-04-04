package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
	"github.com/coderank-dev/coderank/internal/api"
)

// HealthDisplay renders a styled health score breakdown. Scores are
// color-coded: green (≥70), yellow (40–69), red (<40).
func HealthDisplay(h *api.HealthResponse) string {
	var sb strings.Builder

	// Header row: library name + score badge + label
	label := scoreLabel(h.HealthScore)
	scoreBadge := scoreStyle(h.HealthScore).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		Render(fmt.Sprintf("%d/100", h.HealthScore))

	header := lipgloss.JoinHorizontal(lipgloss.Center,
		Title.Bold(true).Render(h.Library),
		"  ",
		scoreBadge,
		"  ",
		label,
	)
	sb.WriteString(header + "\n\n")

	// Build table rows
	categories := []struct {
		name string
		key  string
	}{
		{"Maintenance", "maintenance"},
		{"Security", "security"},
		{"Community", "community"},
		{"Sustainability", "sustainability"},
	}

	var rows [][]string
	for _, cat := range categories {
		score, ok := h.Breakdown[cat.key]
		if !ok {
			continue
		}
		bar := scoreBar(score)
		num := scoreStyle(score).Bold(true).Render(fmt.Sprintf("%3d", score))
		rows = append(rows, []string{cat.name, bar, num})
	}

	if len(rows) > 0 {
		borderStyle := lipgloss.NewStyle().Foreground(BorderClr)

		t := ltable.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(borderStyle).
			BorderColumn(false).
			BorderHeader(true).
			BorderTop(false).
			BorderBottom(false).
			BorderLeft(false).
			BorderRight(false).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == ltable.HeaderRow {
					return lipgloss.NewStyle().
						Foreground(SubtleClr).
						Padding(0, 1)
				}
				switch col {
				case 0: // category name
					return lipgloss.NewStyle().
						Foreground(Muted).
						Width(16).
						Padding(0, 1)
				case 1: // bar
					return lipgloss.NewStyle().Padding(0, 1)
				default: // score
					return lipgloss.NewStyle().
						Align(lipgloss.Right).
						Padding(0, 1)
				}
			}).
			Headers("Category", "Score", "").
			Rows(rows...)

		sb.WriteString(t.Render())
		sb.WriteString("\n")
	}

	// Footer: repo + indexed timestamp
	if h.Repo != "" || h.LastIndexed != "" {
		sb.WriteString("\n")
	}
	if h.Repo != "" {
		sb.WriteString(Subtle.Render(fmt.Sprintf("  repo      github.com/%s", h.Repo)))
		sb.WriteString("\n")
	}
	if h.LastIndexed != "" {
		sb.WriteString(Subtle.Render(fmt.Sprintf("  indexed   %s", h.LastIndexed)))
		sb.WriteString("\n")
	}

	return sb.String()
}

// scoreBar renders a compact visual bar proportional to the score (0–100).
// Uses 15 characters: filled (█) + empty (░).
func scoreBar(score int) string {
	const width = 15
	filled := score * width / 100
	empty := width - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return scoreStyle(score).Render(bar)
}

// scoreLabel returns a styled severity label.
func scoreLabel(score int) string {
	if score >= 70 {
		return Success.Bold(true).Render("● HEALTHY")
	}
	if score >= 40 {
		return Warning.Bold(true).Render("● CAUTION")
	}
	return Error.Bold(true).Render("● LOW")
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
