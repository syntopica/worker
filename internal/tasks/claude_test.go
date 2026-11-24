package tasks

import (
	"reflect"
	"slices"
	"testing"

	"github.com/syntopica/worker/internal/config"
)

func TestClaudeInvocationPassesTheSchemaInline(t *testing.T) {
	t.Parallel()

	call := claude().invoke(&job{profile: &config.TaskProfile{Model: "m"}, prompt: "p", schemaFile: "/s.json", schema: `{"type":"object"}`})
	want := []string{
		"claude", "-p", "--model", "m", "--output-format", "json", "--tools", "", "--strict-mcp-config",
		"--setting-sources", "", "--no-session-persistence", "--json-schema", `{"type":"object"}`,
	}

	if !slices.Equal(call.argv, want) || call.stdin != "p" {
		t.Errorf("argv %q stdin %q", call.argv, call.stdin)
	}
}

func TestClaudeAnswerAndUsage(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		stdout   string
		answer   string
		answered bool
	}{
		"result":            {`{"result": "hi", "is_error": false}`, "hi", true},
		"structured output": {`{"result": "prose", "structured_output": {"a": 1}}`, `{"a": 1}`, true},
		"error envelope":    {`{"result": "hi", "is_error": true}`, "", false},
		"no result":         {`{"result": 3}`, "", false},
		"not JSON":          {"not json", "", false},
	}

	for name, test := range cases {
		if answer, answered := claude().answer(test.stdout); answer != test.answer || answered != test.answered {
			t.Errorf("%s: %q %v", name, answer, answered)
		}
	}

	usage := claude().usage(`{"usage": {"input_tokens": 7, "output_tokens": 2.5}}`)
	if !reflect.DeepEqual(usage, map[string]any{"tokens_in": int64(7)}) {
		t.Errorf("usage %v", usage)
	}
}

func TestClaudeWallSentences(t *testing.T) {
	t.Parallel()

	for _, wall := range []string{
		"Claude AI usage limit reached", "You've hit your weekly limit", "You’ve hit your limit", "5-hour limit reached",
		`{"is_error":true,"api_error_status": 429}`,
	} {
		if !claude().wall.MatchString(wall) {
			t.Errorf("not a wall: %q", wall)
		}
	}

	if claude().wall.MatchString(`"api_error_status": 4290`) {
		t.Error("another status read as a wall")
	}
}
