package jobs

import (
	"time"

	"github.com/dlclark/regexp2/v2"
)

// schemaRegexp adapts the backtracking engine to JSON Schema's boolean matcher.
type schemaRegexp struct {
	compiled *regexp2.Regexp
	source   string
	patterns *schemaRegexes
}

// String returns the source pattern.
func (pattern *schemaRegexp) String() string {
	return pattern.source
}

// MatchString reports whether the pattern matches anywhere in the value.
func (pattern *schemaRegexp) MatchString(value string) bool {
	patterns := pattern.patterns
	if patterns.failed || patterns.remaining <= 0 {
		patterns.failed = true

		return false
	}

	started := time.Now()
	pattern.compiled.MatchTimeout = patterns.remaining
	matched, err := pattern.compiled.MatchString(value)
	patterns.remaining -= time.Since(started)
	patterns.failed = err != nil || patterns.remaining <= 0

	return matched && !patterns.failed
}
