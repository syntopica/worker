package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

func renderReport(out io.Writer, spec reportLayout, raw json.RawMessage, jsonOutput bool) error {
	if spec.rows {
		rows, err := selectReportRows(raw)
		if err != nil {
			return err
		}
		raw = rows
	}
	if jsonOutput {
		return writeJSON(out, raw, true)
	}
	report, err := spec.render(raw)
	if err != nil {
		return err
	}
	if _, err = io.WriteString(out, report); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func selectReportRows(raw json.RawMessage) (json.RawMessage, error) {
	var envelope struct {
		Rows json.RawMessage `json:"rows"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}
	return envelope.Rows, nil
}
