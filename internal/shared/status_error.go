package shared

type AppError struct {
	Status  int
	Message string `json:"message"`
}

func (e AppError) Error() string {
	return e.Message
}

func (e AppError) GetStatus() int {
	return e.Status
}

// func (e *AppError) GetMessage() string {
