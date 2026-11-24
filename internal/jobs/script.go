package jobs

import "embed"

//go:embed sql/*.sql
var scripts embed.FS

// script is the SQL kept under sql/ by that name. A missing name is a
// programming error, so it panics rather than being handled at every call.
func script(name string) string {
	raw, err := scripts.ReadFile("sql/" + name + ".sql")
	if err != nil {
		panic("jobs: no script named " + name)
	}

	return string(raw)
}
