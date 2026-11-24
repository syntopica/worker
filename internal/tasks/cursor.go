package tasks

import "regexp"

var cursorWall = regexp.MustCompile(`(?i)you['’]re out of usage\.\s*switch to auto, or ask your admin to increase your limit` +
	`|you['’]ve hit your usage limit[^\n]{0,300}switch to a different model or set a\s+spend limit`)

// cursor is the cursor-agent print-mode runner in ask mode.
func cursor() *runner {
	return &runner{
		wall:               cursorWall,
		answerRulesOutWall: true,
		invoke:             cursorInvocation,
		// Cursor has no output schema, and a tool-using model says something
		// before the object it was asked for, so the last balanced object in
		// the result is the answer.
		answer: func(stdout string) (string, bool) {
			reply := parseEnvelope(stdout)
			if reply == nil || reply.isError() {
				return "", false
			}

			result, answered := reply.text("result")
			if object := lastJSONObject(result); object != "" {
				return object, true
			}

			return result, answered
		},
		usage: func(stdout string) map[string]any {
			return parseEnvelope(stdout).usage("inputTokens", "outputTokens")
		},
	}
}

// cursorInvocation keeps tools away with --mode ask; --trust grants only the
// workspace. An argv prompt of a few hundred KB makes the CLI exit 0 with no
// output at all, one more reason it goes on stdin.
func cursorInvocation(task *job) invocation {
	argv := []string{task.command("cursor-agent"), "-p", "--trust", "--mode", "ask", outputFormat, jsonFormat}

	if task.profile.Model != "" {
		argv = append(argv, "--model", task.profile.Model)
	}

	return invocation{argv: argv, stdin: task.prompt}
}
