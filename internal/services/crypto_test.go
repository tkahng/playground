package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/services"
)

const validKey = "BlKEVCiNGfRH0gwCVJ6pxxbDW3iLu/nN"

func TestNewCrypto(t *testing.T) {
	t.Parallel()
	t.Run("panics with invalid key length", func(t *testing.T) {
		assert.PanicsWithValue(t, "encryption key must be 32 chars", func() {
			services.NewCrypto("invalid-key-length")
		})
	})

	t.Run("succeeds with valid key length", func(t *testing.T) {
		assert.NotPanics(t, func() {
			services.NewCrypto(validKey)
		})
	})
}

func TestCrypto_EncryptDecrypt(t *testing.T) {
	t.Parallel()
	crypto := services.NewCrypto(validKey)

	t.Run("encrypt and decrypt successfully", func(t *testing.T) {
		originalData := []byte("this is a secret message")

		encrypted, err := crypto.Encrypt(originalData)
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := crypto.Decrypt(encrypted)
		assert.NoError(t, err)
		assert.Equal(t, originalData, decrypted)
	})

	t.Run("decrypt with wrong ciphertext fails", func(t *testing.T) {
		_, err := crypto.Decrypt("invalid-ciphertext")
		assert.Error(t, err)
	})

	t.Run("decrypt with different crypto instance fails", func(t *testing.T) {
		wrongCrypto := services.NewCrypto("another-valid-32-char-secret-key")
		_, err := wrongCrypto.Decrypt("some-encrypted-string")
		assert.Error(t, err)
	})
}
