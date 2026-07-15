package restclient

import (
	"fmt"
)

// Error represents a generic SDK error
type Error struct {
	StatusCode int
	Message    string
	Body       []byte
	Err        error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("SDK error (status %d): %s - %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("SDK error (status %d): %s", e.StatusCode, e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *Error) Unwrap() error {
	return e.Err
}

// NewError creates a new SDK error
func NewError(statusCode int, message string, body []byte, err error) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    message,
		Body:       body,
		Err:        err,
	}
}
