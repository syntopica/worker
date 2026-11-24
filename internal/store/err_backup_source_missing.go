package store

import "errors"

// ErrBackupSourceMissing means the state directory contains no metadata database.
var ErrBackupSourceMissing = errors.New("no metadata store")
