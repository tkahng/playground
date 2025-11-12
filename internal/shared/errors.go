package shared

import (
	"net/http"

	appHttp "github.com/tkahng/playground/internal/tools/http"
)

var (
	ErrUserInfoNotFound = &appHttp.ErrorModel{
		Status: http.StatusUnauthorized,
		Detail: "you are not signed in",
	}
	ErrUserNotFound = &appHttp.ErrorModel{
		Status: http.StatusNotFound,
		Detail: "user not found",
	}
	ErrUserExists = &appHttp.ErrorModel{
		Status: 409,
		Detail: "user already exists",
	}
	ErrInvalidToken = &appHttp.ErrorModel{
		Status: 401,
		Detail: "invalid token",
	}
	ErrTokenExpired = &appHttp.ErrorModel{
		Status: 401,
		Detail: "token expired",
	}
	ErrTokenNotFound = &appHttp.ErrorModel{
		Status: http.StatusNotFound,
		Detail: "token not found",
	}
	ErrPasswordIncorrect = &appHttp.ErrorModel{
		Status: 401,
		Detail: "password is incorrect",
	}
	ErrAccountNotFound = &appHttp.ErrorModel{
		Status: http.StatusNotFound,
		Detail: "account not found",
	}
	ErrAccountProviderConflict = &appHttp.ErrorModel{
		Status: 409,
		Detail: "there is already an account with this provider",
	}
)
