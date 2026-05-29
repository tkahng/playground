package security

import (
	cryptoRand "crypto/rand"
	"math/big"

	"github.com/google/uuid"
)

const DefaultRandomAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// RandomString generates a cryptographically random string with the specified length.
//
// The generated string matches [A-Za-z0-9]+ and it's transparent to URL-encoding.
func RandomString(length int) string {
	return RandomStringWithAlphabet(length, DefaultRandomAlphabet)
}

// RandomStringWithAlphabet generates a cryptographically random string
// with the specified length and characters set.
//
// It panics if for some reason rand.Int returns a non-nil error.
func RandomStringWithAlphabet(length int, alphabet string) string {
	b := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))

	for i := range b {
		n, err := cryptoRand.Int(cryptoRand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = alphabet[n.Int64()]
	}

	return string(b)
}

// PseudorandomString generates a cryptographically random string.
// The name is kept for backwards compatibility; the implementation now uses
// crypto/rand throughout the security package to avoid mixing RNG quality.
func PseudorandomString(length int) string {
	return RandomString(length)
}

// PseudorandomStringWithAlphabet generates a cryptographically random string
// with the specified length and alphabet.
func PseudorandomStringWithAlphabet(length int, alphabet string) string {
	return RandomStringWithAlphabet(length, alphabet)
}

func GenerateTokenKey() string {
	return uuid.NewString()
}
