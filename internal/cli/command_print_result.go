package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (c *command) printResult(id string, result *runResult) (int, error) {
	if result.Control != nil {
		var detail struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(result.Detail, &detail)
		if !refusalCode.MatchString(detail.Error) {
			detail.Error = ""
		}
		_, _ = fmt.Fprintln(c.errOut, strings.TrimSpace(fmt.Sprintf("worker job %s ended %s %s", id, *result.Control, detail.Error)))
		return failedExit, nil
	}

	_, _ = fmt.Fprintf(c.errOut, "worker job %s result %s executor %s\n", id, result.ResultID, formatExecutor(result.Executor))
	var output struct {
		JSON json.RawMessage `json:"json"`
		Text string          `json:"text"`
	}
	if len(result.Output) > 0 {
		if err := json.Unmarshal(result.Output, &output); err != nil {
			return failedExit, fmt.Errorf("decode job output: %w", err)
		}
	}
	if output.JSON != nil && string(output.JSON) != "null" {
		if err := writeJSON(c.out, output.JSON, false); err != nil {
			return failedExit, err
		}
		return 0, nil
	}
	if _, err := fmt.Fprintln(c.out, output.Text); err != nil {
		return failedExit, fmt.Errorf("write job output: %w", err)
	}
	return 0, nil
}
