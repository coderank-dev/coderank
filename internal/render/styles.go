// Package render provides terminal output functions used by all CLI commands.
// Glamour handles markdown rendering with syntax highlighting.
// Lip Gloss handles styled text, borders, and layout.
// No command should use fmt.Println for user-facing output.
package render

import "github.com/charmbracelet/lipgloss"

// Color palette — used consistently across all CLI output.
var (
	Accent    = lipgloss.Color("#22D3EE")
	Green     = lipgloss.Color("#4ADE80")
	Red       = lipgloss.Color("#F87171")
	Yellow    = lipgloss.Color("#FBBF24")
	Blue      = lipgloss.Color("#818CF8")
	Purple    = lipgloss.Color("#C084FC")
	Orange    = lipgloss.Color("#FB923C")
	Muted     = lipgloss.Color("#8888A0")
	SubtleClr = lipgloss.Color("#55556A")
	BorderClr = lipgloss.Color("#22222E")
)

// Reusable styles — every command uses these instead of building ad-hoc styles.
var (
	Title     = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	Heading   = lipgloss.NewStyle().Bold(true)
	Body      = lipgloss.NewStyle()
	Subtle    = lipgloss.NewStyle().Foreground(Muted)
	MutedText = lipgloss.NewStyle().Foreground(Muted)

	Success = lipgloss.NewStyle().Foreground(Green).Bold(true)
	Warning = lipgloss.NewStyle().Foreground(Yellow).Bold(true)
	Error   = lipgloss.NewStyle().Foreground(Red).Bold(true)

	StatusLine = lipgloss.NewStyle().Foreground(Muted).Italic(true)

	DocBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderClr).
		Padding(1, 2)
)
