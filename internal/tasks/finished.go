package tasks

import (
	"strings"

	"github.com/syntopica/worker/internal/jsonanswer"
)

// finished is a runner that exited, with what it left behind.
type finished struct {
	// cli is the runner that ran.
	cli *runner
	// exit is the exit status.
	exit int
	// stdout and stderr are what it printed.
	stdout, stderr string
	// answer is its answer, when answered.
	answer   string
	answered bool
}

// verdict turns the run into an outcome, an error code and an output.
//
// A wall outranks everything; an answer outranks a non-zero exit: codex has
// exited non-zero after writing a usable last message, and the coordinator's
// schema check decides whether that answer is acceptable. An empty answer is
// no answer: agy returns one when a tool it wanted was denied.
func (f *finished) verdict() (how outcome, why reason, output map[string]any) {
	switch {
	case f.wall():
		return failed, quotaWall, nil
	case strings.TrimSpace(f.answer) != "":
		return succeeded, "", map[string]any{"text": f.answer, "json": jsonanswer.Parse(f.answer)}
	case f.exit == 0:
		return failed, noOutput, nil
	}

	messages := f.stderr
	if f.cli.failures != nil {
		messages += "\n" + f.cli.failures(f.stdout)
	}

	return failed, failureCode(messages), nil
}

// usage is what the run measured: the token counts, and for a wall that says
// so, how long it lasts.
func (f *finished) usage(why reason) map[string]any {
	usage := map[string]any{}
	if f.cli.usage != nil {
		usage = f.cli.usage(f.stdout)
	}

	if why != quotaWall {
		return usage
	}

	if seconds, known := wallResetsIn(f.cli.wall, f.judged()); known {
		usage["wall_resets_in_s"] = seconds
	}

	return usage
}

// wall reports whether the runner hit its account's quota rather than
// failing on the task: its wall sentence in the judged text, or a silent
// wall.
func (f *finished) wall() bool {
	if f.cli.answerRulesOutWall && f.answered {
		return false
	}

	if f.cli.wall.MatchString(f.judged()) {
		return true
	}

	return f.cli.silentWall != nil && f.exit == 0 && !f.answered && f.cli.silentWall(f.stdout)
}

// judged is the text a wall is read from: the runner's own where the runner
// can tell it from the model's.
func (f *finished) judged() string {
	if f.cli.judged != nil {
		return f.cli.judged(f.stdout, f.stderr, f.answered)
	}

	return f.stdout + f.stderr
}
