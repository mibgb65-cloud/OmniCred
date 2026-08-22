package credential

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("credential not found")

type ValidationError struct {
	Field   string
	Message string
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", err.Field, err.Message)
}
