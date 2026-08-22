package platform

import (
	"errors"
	"fmt"
)

var (
	ErrAlreadyExists = errors.New("platform already exists")
	ErrInUse         = errors.New("platform is in use")
	ErrNotFound      = errors.New("platform not found")
)

type ValidationError struct {
	Field   string
	Message string
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", err.Field, err.Message)
}
