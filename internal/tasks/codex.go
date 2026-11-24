package tasks

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var codexWall = regexp.MustCompile(`(?i)your workspace is out of credits\.\s*ask your workspace owner to refill`)

// codex is the read-only codex exec runner.
func codex() *runner {
	return &runner{wall: codexWall, invoke: codexInvocation, usage: codexUsage, failures: codexFailures}
}

// codexInvocation asks with -s read-only and no --search: the workspace is
// all it may read. --json prints events on stdout, whose turn.completed
// carries the usage; the answer still comes from the -o file.
func codexInvocation(task *job) invocation {
	answer := filepath.Join(task.scratch, "last-message.txt")
	argv := []string{
		task.command("codex"), "exec", "-", "-C", task.workspace, "-s", "read-only",
		"--skip-git-repo-check", "--json", "-o", answer, "--ephemeral", "--ignore-user-config",
	}

	if task.profile.InputRoot == "" {
		argv = append(argv, "--disable", "shell_tool", "--disable", "shell_snapshot")
	}

	if task.profile.Model != "" {
		argv = append(argv, "-m", task.profile.Model)
	}

	if task.profile.Reasoning != "" {
		argv = append(argv, "-c", "model_reasoning_effort="+task.profile.Reasoning)
	}

	if task.schemaFile != "" {
		argv = append(argv, "--output-schema", task.schemaFile)
	}

	return invocation{argv: argv, stdin: task.prompt, answerFile: answer}
}

// codexEvents is the events of a codex --json stream. Runners can print
// plain warnings among their events, so a line that is not a JSON object is
// skipped rather than failing the whole stream.
func codexEvents(stdout string) []envelope {
	var events []envelope

	for line := range strings.Lines(stdout) {
		if !strings.HasPrefix(line, "{") {
			continue
		}

		if event := parseEnvelope(line); event != nil {
			events = append(events, event)
		}
	}

	return events
}

// codexUsage is the token counts summed over every completed turn.
//
// input_tokens already counts the cached part and output_tokens the
// reasoning part, so the two are taken as they are. A run in which no turn
// completed reports no usage rather than zero.
func codexUsage(stdout string) map[string]any {
	totals := map[string]any{}

	for _, event := range codexEvents(stdout) {
		if kind, _ := event.text("type"); kind != "turn.completed" {
			continue
		}

		usage := event.object("usage")

		for field, key := range map[string]string{"tokens_in": "input_tokens", "tokens_out": "output_tokens"} {
			if count, err := strconv.ParseInt(string(usage[key]), 10, 64); err == nil {
				sum, _ := totals[field].(int64)
				totals[field] = sum + count
			}
		}
	}

	return totals
}

// codexFailures is the error and turn.failed messages, one per line: with
// --json codex moves its own failures from stderr to these events. Agent
// messages and items of type error (configuration warnings) are left out,
// so the model's prose over untrusted input never names a failure.
func codexFailures(stdout string) string {
	var messages []string

	for _, event := range codexEvents(stdout) {
		source := event

		switch kind, _ := event.text("type"); kind {
		case "error":
		case "turn.failed":
			source = event.object("error")
		default:
			continue
		}

		if message, found := source.text("message"); found {
			messages = append(messages, message)
		}
	}

	return strings.Join(messages, "\n")
}
