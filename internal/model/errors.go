package model

import "errors"

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotFound      = errors.New("not found")
	ErrConflict      = errors.New("dependency conflict")
	ErrCycle         = errors.New("dependency cycle")
	ErrAlreadyExists = errors.New("already exists")
	ErrState         = errors.New("invalid state transition")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e *ValidationError) Error() string { return "validation failed" }

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict) || errors.Is(err, ErrAlreadyExists)
}
