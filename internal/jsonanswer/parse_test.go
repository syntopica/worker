package jsonanswer_test

import (
	"testing"

	"github.com/syntopica/worker/internal/jsonanswer"
)

func TestParseToleratesOneSurroundingFence(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`{"a": 1}`:                 `{"a": 1}`,
		"```json\n{\"a\": 1}\n```": `{"a": 1}`,
		"```\n[1]\n```":            `[1]`,
		"  7 ":                     `7`,
		"not json":                 "",
		"":                         "",
	}

	for text, want := range cases {
		if got := string(jsonanswer.Parse(text)); got != want {
			t.Errorf("Parse(%q) = %q, want %q", text, got, want)
		}
	}
}
