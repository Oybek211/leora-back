package errors

import "fmt"

// ErrNotFound indicates missing resource.
type ErrNotFound struct {
	Resource string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

// ErrConflict indicates resource conflict.
type ErrConflict struct {
	Resource string
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("%s conflict", e.Resource)
}
