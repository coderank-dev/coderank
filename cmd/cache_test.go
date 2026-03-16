package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheRequiresLibrariesOrFlag(t *testing.T) {
	err := runCache(cacheCmd, []string{})
	require.Error(t, err, "cache should require libraries or a mode flag")
	assert.Contains(t, err.Error(), "specify libraries",
		"error should tell the user how to use the command")
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		input int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1572864, "1.5 MB"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, formatBytes(tc.input),
			"formatBytes(%d) should return human-readable size", tc.input)
	}
}
