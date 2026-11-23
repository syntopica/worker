package ollama_test

import (
	"encoding/json"
	"testing"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/ollama"
)

// pin is the model every scenario runs.
func pin() *config.ModelPin {
	return &config.ModelPin{Name: "model-a", NumCtx: 40960, KeepAlive: "30m"}
}

// bodyJSON is the request body for an input, as canonical JSON.
func bodyJSON(t *testing.T, input string) string {
	t.Helper()

	body, err := ollama.RequestBody(pin(), json.RawMessage(input))
	if err != nil {
		t.Fatalf("RequestBody(%s): %v", input, err)
	}

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	return string(raw)
}

func TestRequestBodyPinnedOptionsCannotBeOverridden(t *testing.T) {
	t.Parallel()

	got := bodyJSON(t, `{
		"messages": [{"role": "user", "content": "hi"}],
		"options": {"temperature": 0, "num_ctx": 999999, "num_gpu": 0},
		"keep_alive": "24h", "think": true, "stream": false, "format": "json"
	}`)
	want := `{"format":"json","keep_alive":"30m","messages":[{"role":"user","content":"hi"}],` +
		`"model":"model-a","options":{"num_ctx":40960,"temperature":0},"stream":true,"think":false}`

	if got != want {
		t.Errorf("body = %s\nwant   %s", got, want)
	}
}

func TestRequestBodyFormat(t *testing.T) {
	t.Parallel()

	base := `"keep_alive":"30m","messages":[],"model":"model-a","options":{"num_ctx":40960},"stream":true,"think":false}`
	cases := map[string]string{
		`{}`:                               `{` + base,
		`{"messages": null}`:               `{` + base,
		`{"format": "text"}`:               `{` + base,
		`{"schema": {}, "format": "json"}`: `{"format":"json",` + base,
		`{"schema": {"type": "object"}, "format": "json"}`: `{"format":{"type":"object"},` + base,
	}

	for input, want := range cases {
		if got := bodyJSON(t, input); got != want {
			t.Errorf("RequestBody(%s) = %s\nwant %s", input, got, want)
		}
	}
}

func TestRequestBodyRefusesAnInputThatIsNotAnObject(t *testing.T) {
	t.Parallel()

	for _, input := range []string{`[1]`, `"text"`, `{"options": 5}`, ``} {
		if _, err := ollama.RequestBody(pin(), json.RawMessage(input)); err == nil {
			t.Errorf("RequestBody(%q) succeeded", input)
		}
	}
}
