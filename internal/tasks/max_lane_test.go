package tasks

import (
	"reflect"
	"slices"
	"testing"

	"github.com/syntopica/worker/internal/config"
)

func TestMaxLaneInvocation(t *testing.T) {
	t.Parallel()

	call := maxLane().invoke(&job{profile: &config.TaskProfile{Model: "m"}, prompt: "p", schemaFile: "/s.json"})

	if !slices.Equal(call.argv, []string{"max-lane-run", "--model", "m", "--schema", "/s.json"}) || call.stdin != "p" {
		t.Errorf("argv %q stdin %q", call.argv, call.stdin)
	}
}

func TestMaxLaneAnswerAndUsage(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		stdout   string
		answer   string
		answered bool
	}{
		"text":       {`{"answer": "hi"}`, "hi", true},
		"structured": {`{"answer": {"a": 1}}`, `{"a": 1}`, true},
		"no answer":  {`{"answer": null}`, "", false},
		"not JSON":   {"not json", "", false},
	}

	for name, test := range cases {
		if answer, answered := maxLane().answer(test.stdout); answer != test.answer || answered != test.answered {
			t.Errorf("%s: %q %v", name, answer, answered)
		}
	}

	captured := fixture(t, "max_lane_envelope.json")
	if _, answered := maxLane().answer(captured); !answered {
		t.Error("the captured envelope has an answer")
	}

	if got := maxLane().usage(captured); !reflect.DeepEqual(got, map[string]any{"tokens_in": int64(626), "tokens_out": int64(34)}) {
		t.Errorf("usage %v", got)
	}
}

func TestMaxLaneWallSentence(t *testing.T) {
	t.Parallel()

	wall := "max-lane-run: Max lane request failed for keychain service example-service with HTTP status 429."
	if !maxLane().wall.MatchString(wall) {
		t.Error("the rate limit sentence is a wall")
	}
}
