package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthRejectsInvalidKeyFormat(t *testing.T) {
	err := runAuth(authCmd, []string{"invalid_key"})
	assert.ErrorContains(t, err, "cr_sk_",
		"should reject keys that don't start with cr_sk_")
}

func TestAuthRejectsEmptyKey(t *testing.T) {
	err := runAuth(authCmd, []string{"  "})
	assert.ErrorContains(t, err, "cr_sk_",
		"should reject whitespace-only keys")
}
