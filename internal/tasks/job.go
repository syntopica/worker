package tasks

import "github.com/syntopica/worker/internal/config"

// job is what a runner's invocation is built from.
type job struct {
	// profile is the profile the task runs under.
	profile *config.TaskProfile
	// prompt is the task's prompt, credentials redacted.
	prompt string
	// workspace is the directory the runner starts in.
	workspace string
	// scratch is the attempt's private directory, removed when it ends.
	scratch string
	// schemaFile is the path of the output schema; empty when there is none.
	schemaFile string
	// schema is the output schema as JSON text.
	schema string
}

// command is the runner executable: the profile's, else the runner's name.
func (j *job) command(name string) string {
	if j.profile.Command != "" {
		return j.profile.Command
	}

	return name
}
