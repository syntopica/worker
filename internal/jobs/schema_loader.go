package jobs

import "errors"

var errExternalSchema = errors.New("jobs: external schema resources are disabled")

// schemaLoader prevents either filesystem or network reads for job schemas.
type schemaLoader struct{}

// Load refuses any resource the compiler does not already have.
func (schemaLoader) Load(string) (any, error) {
	return nil, errExternalSchema
}
