package tasks

import "testing"

func TestLastJSONObject(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`say {"a": 1} and {"b": {"c": 2}} done`: `{"b": {"c": 2}}`,
		`{broken {"a": 1}`:                      `{"a": 1}`,
		`{"a": 1} trailing {`:                   `{"a": 1}`,
		"no object":                             "",
		"":                                      "",
	}

	for text, want := range cases {
		if got := lastJSONObject(text); got != want {
			t.Errorf("%q: %q", text, got)
		}
	}
}
