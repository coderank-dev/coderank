package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSurfaceCommandRequiresExactlyOneArg(t *testing.T) {
	err := surfaceCmd.Args(surfaceCmd, []string{})
	assert.Error(t, err, "surface requires exactly one argument (library name)")

	err = surfaceCmd.Args(surfaceCmd, []string{"react"})
	assert.NoError(t, err)
}

func TestSurfaceCommandRejectsTooManyArgs(t *testing.T) {
	err := surfaceCmd.Args(surfaceCmd, []string{"react", "nextjs"})
	assert.Error(t, err, "surface should only accept one library at a time")
}
