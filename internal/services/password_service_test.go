//go:build !integration

package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashService_HashAndVerifyPassword(t *testing.T) {
	t.Parallel()
	service := NewHashService()
	password := "mySecretPassword123!"
	hash, err := service.Hash(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	match, err := service.Verify(password, hash)
	assert.NoError(t, err)
	assert.True(t, match)

	// Negative test
	match, err = service.Verify("wrongPassword", hash)
	assert.NoError(t, err)
	assert.False(t, match)
}
