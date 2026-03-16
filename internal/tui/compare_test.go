package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coderank-dev/coderank/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestCompareModelInitialState(t *testing.T) {
	resp := &api.CompareResponse{
		Category: "node orm",
		Libraries: []api.HealthResponse{
			{Library: "prisma", HealthScore: 92, Breakdown: map[string]int{
				"maintenance": 95, "security": 90, "community": 91, "sustainability": 88,
			}},
		},
	}
	model := NewCompareModel(resp)
	assert.Equal(t, "node orm", model.category)
	assert.False(t, model.quitting)
}

func TestCompareModelQuitOnEscape(t *testing.T) {
	resp := &api.CompareResponse{
		Category:  "test",
		Libraries: []api.HealthResponse{},
	}
	model := NewCompareModel(resp)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(CompareModel)
	assert.True(t, m.quitting, "pressing Escape should set quitting to true")
	assert.NotNil(t, cmd, "pressing Escape should return tea.Quit")
}

func TestCompareModelQuitOnQ(t *testing.T) {
	resp := &api.CompareResponse{
		Category:  "test",
		Libraries: []api.HealthResponse{},
	}
	model := NewCompareModel(resp)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m := updated.(CompareModel)
	assert.True(t, m.quitting, "pressing q should set quitting to true")
	assert.NotNil(t, cmd, "pressing q should return tea.Quit")
}
