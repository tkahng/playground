package shared

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenNotFound     = errors.New("token not found")
	ErrPasswordIncorrect = errors.New("password is incorrect")
	ErrAccountNotFound   = errors.New("account not found")
)
