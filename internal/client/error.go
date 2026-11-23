package client

import "fmt"

// Error is a refusal from the coordinator: any non-2xx answer. Transport
// failures are returned as they are, wrapped, and are never an *Error.
type Error struct {
	// Status is the HTTP status.
	Status int
	// Code is the contract's error code, or "unknown" for a body without one.
	Code string
}

// Error implements error.
func (e *Error) Error() string {
	return fmt.Sprintf("worker: %d %s", e.Status, e.Code)
}
