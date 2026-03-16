// Package tui contains Bubble Tea models for interactive CLI commands.
// Only compare and shell use Bubble Tea — everything else is non-interactive.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coderank-dev/coderank/internal/api"
)

// CompareModel is the Bubble Tea model for the compare command.
// It wraps a Bubbles table with health score data.
type CompareModel struct {
	table    table.Model
	category string
	quitting bool
}

// NewCompareModel creates a CompareModel from API response data.
// Columns: Library, Score, Maintenance, Security, Community, Sustainability.
func NewCompareModel(resp *api.CompareResponse) CompareModel {
	columns := []table.Column{
		{Title: "Library", Width: 20},
		{Title: "Score", Width: 8},
		{Title: "Maint.", Width: 8},
		{Title: "Security", Width: 10},
		{Title: "Community", Width: 11},
		{Title: "Sustain.", Width: 10},
	}

	var rows []table.Row
	for _, lib := range resp.Libraries {
		rows = append(rows, table.Row{
			lib.Library,
			fmt.Sprintf("%d", lib.HealthScore),
			fmt.Sprintf("%d", lib.Breakdown["maintenance"]),
			fmt.Sprintf("%d", lib.Breakdown["security"]),
			fmt.Sprintf("%d", lib.Breakdown["community"]),
			fmt.Sprintf("%d", lib.Breakdown["sustainability"]),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)+1, 20)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#22222E")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#22D3EE")).
		Bold(true)
	t.SetStyles(s)

	return CompareModel{
		table:    t,
		category: resp.Category,
	}
}

// Init implements tea.Model.
func (m CompareModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles keyboard input for navigation
// and quitting (q/Esc).
func (m CompareModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View implements tea.Model. Renders the table with a title and help text.
func (m CompareModel) View() string {
	if m.quitting {
		return ""
	}

	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#22D3EE")).
		Render(fmt.Sprintf("Compare: %s", m.category))
	sb.WriteString(title + "\n\n")
	sb.WriteString(m.table.View())
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#8888A0")).
		Render("↑/↓ navigate • q quit"))
	return sb.String()
}
