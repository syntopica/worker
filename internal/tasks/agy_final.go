package tasks

import "strings"

// agyFinal is the terminal result of an agy run.
type agyFinal struct {
	// result is the final result object.
	result envelope
	// streamed is false for the older single-envelope output, still read for
	// rolling upgrades.
	streamed bool
}

// readAgyFinal is the final result in stdout, or nil when it holds no
// complete one.
//
// A stream-json run ends with one {"event":"result","result":{...}} line;
// step updates before it are never an answer, and a truncated stream has no
// such last line. A single JSON object without an event key is the older
// shape.
func readAgyFinal(stdout string) *agyFinal {
	text := strings.TrimSpace(stdout)

	if whole := parseEnvelope(text); whole != nil {
		if _, isEvent := whole["event"]; !isEvent {
			return &agyFinal{result: whole}
		}
	}

	last := parseEnvelope(text[strings.LastIndexByte(text, '\n')+1:])
	if kind, _ := last.text("event"); kind != "result" {
		return nil
	}

	result := last.object("result")
	if result == nil {
		return nil
	}

	return &agyFinal{result: result, streamed: true}
}

// succeeded reports a result that is not a failed stream's.
func (f *agyFinal) succeeded() bool {
	status, _ := f.result.text("status")

	return !f.streamed || status == "SUCCESS"
}

// answer is structured_output first: with a schema, response is prose
// around it. A stream without a final SUCCESS result has no answer.
func (f *agyFinal) answer() (text string, answered bool) {
	if !f.succeeded() {
		return "", false
	}

	if text, answered = f.result.structured("structured_output"); answered {
		return text, true
	}

	if text, answered = f.result.text("response"); answered {
		return text, true
	}

	return f.result.structured("response")
}

// runnerError is the explicit error of a result that did not succeed as a
// stream, on its own line; the rest of a result is model output.
func (f *agyFinal) runnerError() string {
	if f.streamed && f.succeeded() {
		return ""
	}

	if message, found := f.result.text("error"); found {
		return "\n" + message
	}

	return ""
}
