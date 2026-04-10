package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryCommandRequiresAtLeastOneArg(t *testing.T) {
	err := queryCmd.Args(queryCmd, []string{})
	assert.Error(t, err,
		"query should require at least one argument (the query)")
}

func TestQueryCommandAcceptsMultipleWords(t *testing.T) {
	err := queryCmd.Args(queryCmd, []string{"react", "hooks", "state"})
	assert.NoError(t, err,
		"query should accept multiple words that get joined into a query")
}

func TestQueryDefaultMaxTokens(t *testing.T) {
	flag := queryCmd.Flags().Lookup("max-tokens")
	assert.Equal(t, "500", flag.DefValue,
		"default max-tokens should be 500")
}
