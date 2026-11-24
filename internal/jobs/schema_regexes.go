package jobs

import (
	"errors"
	"time"

	"github.com/dlclark/regexp2/v2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var errInvalidSchemaPattern = errors.New("jobs: invalid output schema pattern")

// schemaRegexes shares a backtracking budget across a single output validation.
type schemaRegexes struct {
	remaining time.Duration
	failed    bool
}

func (patterns *schemaRegexes) compile(pattern string) (jsonschema.Regexp, error) {
	normalized, valid := normalizeSchemaRegexp(pattern)
	if !valid {
		return nil, errInvalidSchemaPattern
	}

	expression, err := regexp2.Compile(normalized, regexp2.OptionMaintainCaptureOrder())
	if err != nil {
		return nil, errInvalidSchemaPattern
	}

	return &schemaRegexp{compiled: expression, source: pattern, patterns: patterns}, nil
}
