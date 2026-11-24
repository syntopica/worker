package cli

import (
	"encoding/json"
	"io"
	"os"
)

func (c *command) readJob(path string, stdin bool) (json.RawMessage, error) {
	var (
		raw []byte
		err error
	)
	if stdin && path == "-" {
		raw, err = io.ReadAll(c.in)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, userError("job document cannot be read")
	}
	if !json.Valid(raw) {
		return nil, userError("job document is not valid JSON")
	}
	return json.RawMessage(raw), nil
}
