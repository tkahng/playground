package security

import "github.com/alexedwards/argon2id"

func CreateHash(password string, params *argon2id.Params) (hash string, err error) {
	return argon2id.CreateHash(password, params)
}

func ComparePasswordAndHash(password string, hash string) (match bool, err error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
