package security

import "github.com/alexedwards/argon2id"

func CreateHash(password string, params *argon2id.Params) (hash string, err error) {
	return argon2id.CreateHash(password, params)
}

func ComparePasswordAndHash(password string, hash string) (match bool, err error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MustCreateHash(password string, params *argon2id.Params) string {
	h, e := CreateHash(password, params)
	if e != nil {
		panic(e)
	}
	return h
}

func MustComparePasswordAndHash(password string, hash string) bool {
	a, e := ComparePasswordAndHash(password, hash)
	if e != nil {
		panic(e)
	}
	return a
}
