package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchCommandRequiresAtLeastOneArg(t *testing.T) {
	err := fetchCmd.Args(fetchCmd, []string{})
	assert.Error(t, err,
		"fetch should require at least one argument (the query)")
}

func TestFetchCommandAcceptsMultipleWords(t *testing.T) {
	err := fetchCmd.Args(fetchCmd, []string{"react", "hooks", "state"})
	assert.NoError(t, err,
		"fetch should accept multiple words that get joined into a query")
}

func TestFetchDefaultMaxTokens(t *testing.T) {
	flag := fetchCmd.Flags().Lookup("max-tokens")
	assert.Equal(t, "5000", flag.DefValue,
		"default max-tokens should be 5000")
}
