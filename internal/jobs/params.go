package jobs

import "database/sql"

// params are the named parameters of one script; a script uses the ones it
// names and ignores the rest.
type params map[string]any

func (p params) list() []any {
	out := make([]any, 0, len(p))
	for name, value := range p {
		out = append(out, sql.Named(name, value))
	}

	return out
}
