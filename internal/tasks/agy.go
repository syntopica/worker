package tasks

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

// agyPreface opens every agy prompt: a denied tool request ends the turn
// with an empty response.
const agyPreface = "Answer in plain text from this prompt alone. Do not use any tool, run any command " +
	"or read any file: everything you need is below.\n\n"

var agyWall = regexp.MustCompile(`(?i)individual quota reached\.\s*please upgrade your subscription`)

// agy is the agy print-mode runner in its no-tools posture.
func agy() *runner {
	return &runner{
		wall:   agyWall,
		invoke: agyInvocation,
		answer: func(stdout string) (string, bool) {
			if final := readAgyFinal(stdout); final != nil {
				return final.answer()
			}

			return "", false
		},
		// Only the final result's usage counts: step updates are never summed.
		usage: func(stdout string) map[string]any {
			if final := readAgyFinal(stdout); final != nil {
				return final.result.usage("input_tokens", "output_tokens")
			}

			return map[string]any{}
		},
		// A successful result and every step update are model output over
		// untrusted input and may quote a wall sentence, so stdout
		// contributes only the error of a failed final result.
		judged: func(stdout, stderr string, answered bool) string {
			if final := readAgyFinal(stdout); final != nil && !answered {
				return stderr + final.runnerError()
			}

			return stderr
		},
		// agy exits 0 with no answer when its quota is spent, in the older
		// output shape only: an incomplete stream is not quota evidence.
		silentWall: func(stdout string) bool {
			final := readAgyFinal(stdout)

			return final != nil && !final.streamed
		},
	}
}

// agyInvocation asks with --sandbox --mode plan --disable-slash-commands and
// never skips permissions, so a tool request blocks rather than being
// approved; a task needs none because everything it may consider is in the
// prompt. The prompt travels as the one user event of stream-json input,
// escaped onto a single ASCII line.
func agyInvocation(task *job) invocation {
	minutes := max(1, int(task.profile.Timeout/time.Minute))
	argv := []string{
		task.command("agy"), "--input-format", "stream-json", outputFormat, "stream-json",
		"--print-timeout", strconv.Itoa(minutes) + "m", "--sandbox", "--mode", "plan", "--disable-slash-commands",
	}

	if task.profile.Model != "" {
		argv = append(argv, "--model", task.profile.Model)
	}

	if task.schemaFile != "" {
		argv = append(argv, "--json-schema", task.schemaFile)
	}

	content, _ := json.Marshal(agyPreface + task.prompt) //nolint:errchkjson // a string always marshals.

	return invocation{argv: argv, stdin: `{"event": "user", "message": {"content": ` + asciiOnly(content) + "}}\n"}
}

// asciiOnly writes JSON text with every non-ASCII character escaped.
func asciiOnly(text []byte) string {
	var ascii strings.Builder

	for _, char := range string(text) {
		switch {
		case char < utf8.RuneSelf:
			ascii.WriteRune(char)
		case char > 0xFFFF: //nolint:mnd // past the basic plane: a surrogate pair.
			high, low := utf16.EncodeRune(char)
			fmt.Fprintf(&ascii, `\u%04x\u%04x`, high, low)
		default:
			fmt.Fprintf(&ascii, `\u%04x`, char)
		}
	}

	return ascii.String()
}
