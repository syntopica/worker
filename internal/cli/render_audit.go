package cli

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

func renderAudit(raw json.RawMessage) (string, error) {
	var rows []struct {
		Time      float64 `json:"time"`
		Action    string  `json:"action"`
		Class     string  `json:"class"`
		JobID     string  `json:"job_id"`
		Principal string  `json:"principal"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return "", fmt.Errorf("read audit rows: %w", err)
	}

	var out strings.Builder
	for _, row := range rows {
		when := time.Unix(int64(math.Floor(row.Time)), 0).UTC().Format(time.DateTime)
		fmt.Fprintf(&out, "%s %-6s %-8s %s %s\n", when, row.Action, row.Class, row.JobID, row.Principal)
	}
	return out.String(), nil
}
