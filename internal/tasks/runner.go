package tasks

import "regexp"

// runner is what the loop knows about one runner CLI.
type runner struct {
	// wall matches the sentence the runner prints when its account's quota
	// is spent.
	wall *regexp.Regexp
	// answerRulesOutWall is set when the runner's stdout carries the model's
	// answer, which may quote the wall sentence from the material it was
	// given: a run that answered is then never a wall.
	answerRulesOutWall bool
	// invoke builds the call for one task.
	invoke func(task *job) invocation
	// answer reads the answer from stdout; unused when the invocation names
	// an answer file.
	answer func(stdout string) (text string, answered bool)
	// usage reads the token counts from stdout.
	usage func(stdout string) map[string]any
	// judged is the text a wall is looked for in; nil means all the output.
	judged func(stdout, stderr string, answered bool) string
	// silentWall reports whether an exit 0 without an answer is a spent
	// quota; nil means it never is.
	silentWall func(stdout string) bool
	// failures is the runner's own failure messages in stdout, beyond its
	// stderr; nil means it has none there.
	failures func(stdout string) string
}

// runnerFor is the named runner. The producer's runner was checked at submit
// and the profile's at load, so any other name is called as cursor is.
func runnerFor(name string) *runner {
	switch name {
	case "codex":
		return codex()
	case "agy":
		return agy()
	case "claude":
		return claude()
	case "max-lane":
		return maxLane()
	default:
		return cursor()
	}
}
