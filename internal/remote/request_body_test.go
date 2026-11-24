package remote

import (
	"encoding/json"
	"testing"
)

// encoded is a request body as the endpoint receives it.
func encoded(t *testing.T, model, input string, zdr bool, fallbacks ...string) map[string]any {
	t.Helper()

	body, err := RequestBody(model, json.RawMessage(input), zdr, fallbacks)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	var sent map[string]any
	if err = json.Unmarshal(raw, &sent); err != nil {
		t.Fatal(err)
	}

	return sent
}

func TestTheBodyCarriesAStrictSchemaAndZDRWhenAsked(t *testing.T) {
	t.Parallel()

	job := `{"messages": [{"role": "user", "content": "hi"}],` +
		` "schema": {"type": "object"}, "options": {"temperature": 0, "seed": 4}}`
	body := encoded(t, "vendor/m:free", job, true)

	format := as[map[string]any](t, as[map[string]any](t, body["response_format"])["json_schema"])
	if format["strict"] != true || as[map[string]any](t, format["schema"])["type"] != "object" {
		t.Errorf("format %v", format)
	}

	if as[map[string]any](t, body["provider"])["zdr"] != true || body["temperature"] != 0.0 {
		t.Errorf("body %v", body)
	}

	for _, absent := range []string{"models", "seed"} {
		if _, set := body[absent]; set {
			t.Errorf("%s set", absent)
		}
	}

	plain := encoded(t, "vendor/m:free", `{"messages": [], "schema": null}`, false)
	for _, absent := range []string{"provider", "response_format", "temperature"} {
		if _, set := plain[absent]; set {
			t.Errorf("%s set", absent)
		}
	}
}

func TestFallbacksBecomeTheOrderedModelsList(t *testing.T) {
	t.Parallel()

	body := encoded(t, "vendor/a:free", `{"messages": []}`, false, "vendor/b:free")

	models := as[[]any](t, body["models"])
	if len(models) != 2 || models[0] != "vendor/a:free" || models[1] != "vendor/b:free" {
		t.Errorf("models %v", models)
	}
}

func TestCredentialsNeverLeaveInTheRemoteBody(t *testing.T) {
	t.Parallel()

	job := `{"messages": [{"role": "user", "content": "Your code is 482913"}, {"role": "user"},` +
		` {"role": "user", "content": ["token=abcdef1234"]}]}`
	messages := as[[]any](t, encoded(t, "vendor/model:free", job, false)["messages"])

	first := as[map[string]any](t, messages[0])
	if first["content"] != "Your code is [REDACTED]" || first["role"] != "user" {
		t.Errorf("first %v", first)
	}

	if as[map[string]any](t, messages[1])["content"] != "" {
		t.Errorf("second %v", messages[1])
	}

	if as[map[string]any](t, messages[2])["content"] != `["token=[REDACTED]"]` {
		t.Errorf("third %v", messages[2])
	}
}

func TestAnInputWithoutMessagesIsRefused(t *testing.T) {
	t.Parallel()

	for _, input := range []string{`{}`, `not json`} {
		if _, err := RequestBody("m", json.RawMessage(input), false, nil); err == nil {
			t.Errorf("%s accepted", input)
		}
	}
}
