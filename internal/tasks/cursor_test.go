package tasks

import (
	"reflect"
	"slices"
	"testing"

	"github.com/syntopica/worker/internal/config"
)

func TestCursorInvocation(t *testing.T) {
	t.Parallel()

	call := cursor().invoke(&job{profile: &config.TaskProfile{Model: "m"}, prompt: "p"})
	want := []string{"cursor-agent", "-p", "--trust", "--mode", "ask", "--output-format", "json", "--model", "m"}

	if !slices.Equal(call.argv, want) || call.stdin != "p" {
		t.Errorf("argv %q stdin %q", call.argv, call.stdin)
	}
}

func TestCursorAnswerIsTheLastObjectOfANarratedResult(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		stdout   string
		answer   string
		answered bool
	}{
		"narrated":       {`{"result": "First {\"a\": 1} then {\"b\": 2}."}`, `{"b": 2}`, true},
		"prose":          {`{"result": "no object"}`, "no object", true},
		"error envelope": {`{"result": "x", "is_error": true}`, "", false},
		"no result":      {`{"result": null}`, "", false},
		"not JSON":       {"not json", "", false},
	}

	for name, test := range cases {
		if answer, answered := cursor().answer(test.stdout); answer != test.answer || answered != test.answered {
			t.Errorf("%s: %q %v", name, answer, answered)
		}
	}
}

func TestCursorUsageMapsTheCapturedEnvelope(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		stdout string
		want   map[string]any
	}{
		"captured":         {fixture(t, "cursor_envelope.json"), map[string]any{"tokens_in": int64(17584), "tokens_out": int64(54)}},
		"not JSON":         {"not json", map[string]any{}},
		"a count in words": {`{"usage": {"inputTokens": "many"}}`, map[string]any{}},
	}

	for name, test := range cases {
		if got := cursor().usage(test.stdout); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: %v", name, got)
		}
	}
}

func TestCursorWallSentences(t *testing.T) {
	t.Parallel()

	for _, wall := range []string{
		"You're out of usage. Switch to Auto, or ask your admin to increase your limit",
		"You’ve hit your usage limit for this model. Switch to a different model or set a\nSpend Limit",
	} {
		if !cursor().wall.MatchString(wall) {
			t.Errorf("not a wall: %q", wall)
		}
	}
}
