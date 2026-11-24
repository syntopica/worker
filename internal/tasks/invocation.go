package tasks

// invocation is how one runner is called for one task. The prompt is never
// an argument: argv is visible to every local process, and a prompt that
// starts with a dash or equals a subcommand name would be parsed as one.
type invocation struct {
	// argv is the command line.
	argv []string
	// stdin is what the runner reads: the prompt, in the runner's shape.
	stdin string
	// answerFile is where the runner writes its answer; empty means the
	// answer is in its stdout.
	answerFile string
}
