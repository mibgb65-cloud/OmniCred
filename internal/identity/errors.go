package identity

import "errors"

var ErrNotFound = errors.New("identity profile not found")

type ValidationError struct {
	Field   string
	Message string
}

func (err *ValidationError) Error() string {
	return err.Field + " " + err.Message
}
