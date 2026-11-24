package store

import "errors"

// ErrBackupDestination means the requested backup destination could not be safely created.
var ErrBackupDestination = errors.New("backup destination unavailable")
