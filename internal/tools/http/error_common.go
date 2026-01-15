package http

import (
	"net/http"
)

var (
	ErrUserInfoNotFound = &ErrorModel{
		Status: http.StatusUnauthorized,
		Detail: "you are not signed in",
	}
	ErrUserNotFound = &ErrorModel{
		Status: http.StatusNotFound,
		Detail: "user not found",
	}
	ErrUserExists = &ErrorModel{
		Status: 409,
		Detail: "user already exists",
	}
	ErrInvalidToken = &ErrorModel{
		Status: 401,
		Detail: "invalid token",
	}
	ErrTokenExpired = &ErrorModel{
		Status: 401,
		Detail: "token expired",
	}
	ErrTokenNotFound = &ErrorModel{
		Status: http.StatusNotFound,
		Detail: "token not found",
	}
	ErrPasswordIncorrect = &ErrorModel{
		Status: 401,
		Detail: "password is incorrect",
	}
	ErrAccountNotFound = &ErrorModel{
		Status: http.StatusNotFound,
		Detail: "account not found",
	}
	ErrAccountProviderConflict = &ErrorModel{
		Status: 409,
		Detail: "account provider conflict",
	}
)
