package shared

type AppError struct {
	status  int
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) GetStatus() int {
	return e.status
}

var (
	ErrUserNotFound = &AppError{
		status:  404,
		Message: "user not found",
	}
	ErrUserExists = &AppError{
		status:  409,
		Message: "user already exists",
	}
	ErrInvalidToken = &AppError{
		status:  401,
		Message: "invalid token",
	}
	ErrTokenExpired = &AppError{
		status:  401,
		Message: "token expired",
	}
	ErrTokenNotFound = &AppError{
		status:  404,
		Message: "token not found",
	}
	ErrPasswordIncorrect = &AppError{
		status:  401,
		Message: "password is incorrect",
	}
	ErrAccountNotFound = &AppError{
		status:  404,
		Message: "account not found",
	}
)
