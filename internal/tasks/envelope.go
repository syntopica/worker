package tasks

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// envelope is a JSON object a runner printed, its values still undecoded.
type envelope map[string]json.RawMessage

// parseEnvelope is the object in text, or nil when text is not one.
func parseEnvelope(text string) envelope {
	var parsed envelope
	if json.Unmarshal([]byte(text), &parsed) != nil {
		return nil
	}

	return parsed
}

// text is a string member.
func (e envelope) text(key string) (value string, found bool) {
	if raw := bytes.TrimSpace(e[key]); len(raw) == 0 || raw[0] != '"' {
		return "", false
	}

	return value, json.Unmarshal(e[key], &value) == nil
}

// structured is an object or array member, as the JSON the runner wrote.
func (e envelope) structured(key string) (value string, found bool) {
	raw := bytes.TrimSpace(e[key])
	if len(raw) == 0 || (raw[0] != '{' && raw[0] != '[') {
		return "", false
	}

	return string(raw), true
}

// object is an object member, or nil.
func (e envelope) object(key string) envelope {
	return parseEnvelope(string(e[key]))
}

// isError reports an envelope that says its run failed.
func (e envelope) isError() bool {
	return string(bytes.TrimSpace(e["is_error"])) == "true"
}

// usage is tokens_in and tokens_out where the envelope's usage object
// carries them as integers. A missing usage or a non-integer count is left
// out rather than guessed, so the ledger stores null instead of a wrong
// number. No runner envelope carries a price.
func (e envelope) usage(inKey, outKey string) map[string]any {
	counts := map[string]any{}
	usage := e.object("usage")

	for field, key := range map[string]string{"tokens_in": inKey, "tokens_out": outKey} {
		if count, err := strconv.ParseInt(string(bytes.TrimSpace(usage[key])), 10, 64); err == nil {
			counts[field] = count
		}
	}

	return counts
}
