package render

import (
	"testing"

	"github.com/coderank-dev/coderank/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestHealthDisplayShowsAllCategories(t *testing.T) {
	h := &api.HealthResponse{
		Library:     "react",
		HealthScore: 91,
		Breakdown: map[string]int{
			"maintenance":    95,
			"security":       88,
			"community":      94,
			"sustainability": 85,
		},
	}

	output := HealthDisplay(h)
	assert.Contains(t, output, "react")
	assert.Contains(t, output, "91")
	assert.Contains(t, output, "Maintenance")
	assert.Contains(t, output, "Security")
	assert.Contains(t, output, "Community")
	assert.Contains(t, output, "Sustainability")
	assert.Contains(t, output, "HEALTHY",
		"score ≥70 should display HEALTHY label")
}

func TestScoreLabelReturnsCorrectSeverity(t *testing.T) {
	assert.Contains(t, scoreLabel(85), "HEALTHY",
		"score ≥70 should be HEALTHY")
	assert.Contains(t, scoreLabel(70), "HEALTHY",
		"score of exactly 70 should be HEALTHY")
	assert.Contains(t, scoreLabel(55), "CAUTION",
		"score 40–69 should be CAUTION")
	assert.Contains(t, scoreLabel(40), "CAUTION",
		"score of exactly 40 should be CAUTION")
	assert.Contains(t, scoreLabel(25), "LOW",
		"score <40 should be LOW")
	assert.Contains(t, scoreLabel(0), "LOW",
		"score of 0 should be LOW")
}

func TestScoreBarLength(t *testing.T) {
	// Strip ANSI escape codes by checking raw rune count isn't empty
	bar100 := scoreBar(100)
	bar0 := scoreBar(0)
	assert.NotEmpty(t, bar100, "score bar should not be empty for score 100")
	assert.NotEmpty(t, bar0, "score bar should not be empty for score 0")
}
