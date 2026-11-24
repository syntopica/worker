package api

import (
	"net/url"
	"strings"
)

// query keeps the last nonempty value for each key, as in contract v1.
type query map[string]string

func parseQuery(raw string) query {
	values := make(query)
	for pair := range strings.SplitSeq(raw, "&") {
		key, value, _ := strings.Cut(pair, "=")
		key, keyErr := url.QueryUnescape(key)
		value, valueErr := url.QueryUnescape(value)
		if keyErr == nil && valueErr == nil && value != "" {
			values[key] = value
		}
	}
	return values
}

func (q query) optional(key string) *string {
	value, found := q[key]
	if !found {
		return nil
	}
	return &value
}

func (q query) integer(key string, fallback int64) (int64, error) {
	value := q.optional(key)
	if value == nil {
		return fallback, nil
	}
	return integerValue(strings.TrimSpace(*value))
}

func (q query) number(key string, fallback float64) (float64, error) {
	value := q.optional(key)
	if value == nil {
		return fallback, nil
	}
	return numberValue(*value)
}
