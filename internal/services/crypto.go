package services

import (
	"github.com/tkahng/playground/internal/tools/security"
)

type Crypto struct {
	encryptionKey string // 32 char AES key
}

type Encryptor interface {
	Encrypt(data []byte) (string, error)
	Decrypt(cipherText string) ([]byte, error)
}

func NewCrypto(encryptionKey string) *Crypto {
	if len(encryptionKey) != 32 {
		panic("encryption key must be 32 chars")
	}
	return &Crypto{encryptionKey: encryptionKey}
}

func (c *Crypto) Encrypt(data []byte) (string, error) {
	return security.Encrypt(data, c.encryptionKey)
}

func (c *Crypto) Decrypt(cipherText string) ([]byte, error) {
	return security.Decrypt(cipherText, c.encryptionKey)
}
