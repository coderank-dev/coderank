// Package render provides shared Lip Gloss styles for the coderank CLI.
package render

import "github.com/charmbracelet/lipgloss"

var (
	Title   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	Warning = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	Error   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	Subtle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	Border  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)
