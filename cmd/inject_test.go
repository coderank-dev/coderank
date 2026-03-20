package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInjectCommandAcceptsNoArgs(t *testing.T) {
	// With no args, inject reads from .coderank.yml — should not error on arg count
	err := injectCmd.Args(injectCmd, []string{})
	assert.NoError(t, err,
		"inject with no args should be valid (reads from config)")
}

func TestInjectCommandAcceptsLibraryArgs(t *testing.T) {
	err := injectCmd.Args(injectCmd, []string{"react", "nextjs", "prisma"})
	assert.NoError(t, err,
		"inject should accept library names as positional arguments")
}
