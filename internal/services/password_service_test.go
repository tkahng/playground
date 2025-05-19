package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordService_HashAndVerifyPassword(t *testing.T) {
	service := NewPasswordService()
	password := "mySecretPassword123!"
	hash, err := service.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	match, err := service.VerifyPassword(hash, password)
	assert.NoError(t, err)
	assert.True(t, match)

	// Negative test
	match, err = service.VerifyPassword(hash, "wrongPassword")
	assert.NoError(t, err)
	assert.False(t, match)
}
