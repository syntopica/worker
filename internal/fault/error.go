package fault

import "fmt"

// Error is a refusal with its HTTP status and a code from the contract's
// fixed vocabulary; it never carries content.
type Error struct {
	Status int
	Code   string
}

// New refuses with a status and a code.
func New(status int, code string) *Error {
	return &Error{Status: status, Code: code}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d %s", e.Status, e.Code)
}
