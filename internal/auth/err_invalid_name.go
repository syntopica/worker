package auth

import "errors"

// ErrInvalidName refuses a name that could not safely be a file name.
var ErrInvalidName = errors.New("invalid principal name")
