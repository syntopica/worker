package tasks

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
)

const (
	agyWallText = "Individual quota reached. Please upgrade your subscription. Resets in 1h."
	agySteps    = `{"event": "init"}` + "\n" + `{"event": "step_update", "text": "partial"}` + "\n"
)

// fixture is a captured runner output.
func fixture(t *testing.T, name string) string {
	t.Helper()

	text, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}

	return string(text)
}

// agyResult is a final stream line whose result is the given object.
func agyResult(result string) string {
	return `{"event": "result", "result": ` + result + "}\n"
}

func TestAgyPromptIsOnlyOnStdinAsOneEscapedLine(t *testing.T) {
	t.Parallel()

	prompt := "--help\n--version   é \U0001F600 \"quoted\" --print=x"
	call := agy().invoke(&job{
		profile:    &config.TaskProfile{Model: "m", Timeout: 10 * time.Minute},
		prompt:     prompt,
		schemaFile: "/s.json",
	})

	want := []string{
		"agy", "--input-format", "stream-json", "--output-format", "stream-json", "--print-timeout", "10m",
		"--sandbox", "--mode", "plan", "--disable-slash-commands", "--model", "m", "--json-schema", "/s.json",
	}
	if !slices.Equal(call.argv, want) {
		t.Errorf("argv %q", call.argv)
	}

	if !strings.HasSuffix(call.stdin, "\n") || strings.Count(call.stdin, "\n") != 1 {
		t.Errorf("stdin is not one line: %q", call.stdin)
	}

	if strings.ContainsFunc(call.stdin, func(char rune) bool { return char > 127 }) {
		t.Errorf("stdin is not ASCII: %q", call.stdin)
	}

	var event struct {
		Event   string `json:"event"`
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal([]byte(call.stdin), &event); err != nil {
		t.Fatal(err)
	}

	if event.Event != "user" || event.Message.Content != agyPreface+prompt {
		t.Errorf("event %+v", event)
	}
}

func TestAgyAnswer(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		stdout   string
		answer   string
		answered bool
	}{
		"structured output beats prose": {
			agySteps + agyResult(`{"status": "SUCCESS", "response": "prose", "structured_output": {"a": 1}}`), `{"a": 1}`, true,
		},
		"final response":              {agyResult(`{"status": "SUCCESS", "response": "hi"}`), "hi", true},
		"structured response":         {agyResult(`{"status": "SUCCESS", "response": [1]}`), "[1]", true},
		"legacy response":             {`{"response": "old"}`, "old", true},
		"legacy structured output":    {`{"structured_output": [1]}`, "[1]", true},
		"line separator in answer":    {agyResult("{\"status\": \"SUCCESS\", \"response\": \"line answer\"}"), "line answer", true},
		"truncated stream":            {agySteps, "", false},
		"malformed last line":         {agySteps + `{"event":"result","res`, "", false},
		"empty":                       {"", "", false},
		"not JSON":                    {"not json", "", false},
		"terminal error":              {agyResult(`{"status": "ERROR", "response": "x"}`), "", false},
		"result that is no object":    {`{"event": "result", "result": 3}`, "", false},
		"legacy without an answer":    {`{"usage": {}}`, "", false},
		"a lone array is no envelope": {"[1]", "", false},
	}

	for name, test := range cases {
		if answer, answered := agy().answer(test.stdout); answer != test.answer || answered != test.answered {
			t.Errorf("%s: %q %v", name, answer, answered)
		}
	}
}

func TestAgyUsageIsTheFinalCumulativeResultOnly(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		stdout string
		want   map[string]any
	}{
		"stream": {
			`{"event": "step_update", "usage": {"input_tokens": 5, "output_tokens": 5}}` + "\n" +
				agyResult(`{"status": "SUCCESS", "usage": {"input_tokens": 1, "output_tokens": 2}}`),
			map[string]any{"tokens_in": int64(1), "tokens_out": int64(2)},
		},
		"no result":        {`{"event": "init"}`, map[string]any{}},
		"legacy":           {`{"usage": {"input_tokens": 3, "output_tokens": 4}}`, map[string]any{"tokens_in": int64(3), "tokens_out": int64(4)}},
		"captured error":   {fixture(t, "agy_error_envelope.json"), map[string]any{"tokens_in": int64(0), "tokens_out": int64(0)}},
		"captured success": {fixture(t, "agy_success_envelope.json"), map[string]any{"tokens_in": int64(2214), "tokens_out": int64(19)}},
		"a boolean count":  {`{"usage": {"input_tokens": true}}`, map[string]any{}},
		"a lone array":     {"[1]", map[string]any{}},
	}

	for name, test := range cases {
		if got := agy().usage(test.stdout); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: %v", name, got)
		}
	}
}

func TestAgyJudgesOnlyTheRunnersOwnText(t *testing.T) {
	t.Parallel()

	quoted, err := json.Marshal(agyWallText)
	if err != nil {
		t.Fatal(err)
	}

	wall := string(quoted)

	cases := map[string]struct {
		stdout, stderr string
		answered       bool
		want           string
	}{
		"prose quoting the wall": {
			`{"event": "step_update", "text": ` + wall + "}\n" + agyResult(`{"status": "SUCCESS", "response": `+wall+`}`), "", true, "",
		},
		"truncated stream":             {`{"event": "step_update", "text": ` + wall + "}\n", "", false, ""},
		"terminal error":               {agyResult(`{"status": "ERROR", "error": ` + wall + `}`), "", false, "\n" + agyWallText},
		"terminal error's response":    {agyResult(`{"status": "ERROR", "response": ` + wall + `}`), "", false, ""},
		"stderr":                       {"", agyWallText, false, agyWallText},
		"legacy unanswered envelope":   {`{"response": "", "error": ` + wall + `}`, "", false, "\n" + agyWallText},
		"legacy answered envelope":     {`{"response": ` + wall + `}`, "", true, ""},
		"unanswered successful stream": {agyResult(`{"status": "SUCCESS", "error": ` + wall + `}`), "", false, ""},
	}

	for name, test := range cases {
		if got := agy().judged(test.stdout, test.stderr, test.answered); got != test.want {
			t.Errorf("%s: %q", name, got)
		}
	}

	if seconds, known := wallResetsIn(agy().wall, "\n"+agyWallText); !known || seconds != 3600 {
		t.Errorf("reset %v %v", seconds, known)
	}
}

func TestAgySilentWallIsTheOlderShapeOnly(t *testing.T) {
	t.Parallel()

	if !agy().silentWall(`{"response": null}`) {
		t.Error("an older envelope without an answer is a silent wall")
	}

	if agy().silentWall(agySteps) || agy().silentWall(agyResult(`{"status": "ERROR"}`)) {
		t.Error("a stream is never a silent wall")
	}
}
