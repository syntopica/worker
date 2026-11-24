package tasks

import "regexp"

var maxLaneWall = regexp.MustCompile(`(?i)max-lane-run: max lane request failed for keychain service ` +
	`[^\n]{1,80} with http status 429\.`)

// maxLane is the max-lane-run runner: one forced-tool request, no tools and
// no workspace, so everything the model may consider is in the prompt.
func maxLane() *runner {
	return &runner{
		wall:               maxLaneWall,
		answerRulesOutWall: true,
		invoke: func(task *job) invocation {
			argv := []string{task.command("max-lane-run"), "--model", task.profile.Model}

			if task.schemaFile != "" {
				argv = append(argv, "--schema", task.schemaFile)
			}

			return invocation{argv: argv, stdin: task.prompt}
		},
		// Free text as it is, a structured answer as JSON.
		answer: func(stdout string) (string, bool) {
			reply := parseEnvelope(stdout)

			if text, answered := reply.text("answer"); answered {
				return text, true
			}

			return reply.structured("answer")
		},
		usage: func(stdout string) map[string]any {
			return parseEnvelope(stdout).usage("input_tokens", "output_tokens")
		},
	}
}
