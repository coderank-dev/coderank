package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommandHasCorrectName(t *testing.T) {
	assert.Equal(t, "coderank", rootCmd.Use,
		"binary name should be 'coderank'")
	assert.NotEmpty(t, rootCmd.Short)
}

func TestRootCommandExecutesWithoutConfigFile(t *testing.T) {
	// First-time users won't have .coderank.yml — must not error
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}
