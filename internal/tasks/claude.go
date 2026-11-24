package tasks

import "regexp"

var claudeWall = regexp.MustCompile(`(?i)claude ai usage limit reached|you['’]ve hit your (?:\w+ )?limit` +
	`|(?:5-hour|session|weekly|opus|sonnet) limit reached|"api_error_status":\s*429\b`)

// claude is the Claude Code CLI print-mode runner: no tools, no settings,
// no session file.
func claude() *runner {
	return &runner{
		wall:               claudeWall,
		answerRulesOutWall: true,
		invoke:             claudeInvocation,
		// structured_output as JSON, else result; none for an error envelope.
		answer: func(stdout string) (string, bool) {
			reply := parseEnvelope(stdout)
			if reply == nil || reply.isError() {
				return "", false
			}

			if text, answered := reply.structured("structured_output"); answered {
				return text, true
			}

			return reply.text("result")
		},
		usage: func(stdout string) map[string]any {
			return parseEnvelope(stdout).usage("input_tokens", "output_tokens")
		},
	}
}

// claudeInvocation passes the schema inline, as the CLI expects. No settings
// means no hooks and no plugins, and no session file means the run never
// lands in a conversation archive. The account is the one the profile's
// environment selects.
func claudeInvocation(task *job) invocation {
	argv := []string{
		task.command("claude"), "-p", "--model", task.profile.Model, outputFormat, jsonFormat,
		"--tools", "", "--strict-mcp-config", "--setting-sources", "", "--no-session-persistence",
	}

	if task.schemaFile != "" {
		argv = append(argv, "--json-schema", task.schema)
	}

	return invocation{argv: argv, stdin: task.prompt}
}
