package ollama_test

import (
	"encoding/json"
	"testing"

	"github.com/syntopica/worker/internal/ollama"
)

func TestAnswerOutput(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`{"label": "x"}`: `{"json":{"label":"x"},"text":"{\"label\": \"x\"}"}`,
		"plain words":    `{"json":null,"text":"plain words"}`,
		"":               `{"json":null,"text":""}`,
	}

	for content, want := range cases {
		got, err := json.Marshal(ollama.Answer{Content: content}.Output())
		if err != nil || string(got) != want {
			t.Errorf("Output(%q) = %s, %v, want %s", content, got, err, want)
		}
	}
}
