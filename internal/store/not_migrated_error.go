package store

import "fmt"

// NotMigratedError refuses a store opened before `worker serve` migrated it.
type NotMigratedError struct {
	Found  int
	Wanted int
}

func (e *NotMigratedError) Error() string {
	return fmt.Sprintf("store schema version %d is older than %d; run worker serve", e.Found, e.Wanted)
}
