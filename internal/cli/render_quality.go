package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

type qualityExecutor struct {
	Queue    string `json:"queue"`
	Tier     string `json:"tier"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

func renderQuality(raw json.RawMessage) (string, error) {
	var report struct {
		Attempts []struct {
			qualityExecutor

			Attempts         int64   `json:"attempts"`
			Succeeded        int64   `json:"succeeded"`
			SchemaViolations int64   `json:"schema_violations"`
			Failed           int64   `json:"failed"`
			Preempted        int64   `json:"preempted"`
			MeanWall         float64 `json:"mean_wall_s"`
		} `json:"attempts"`
		Ratings []struct {
			qualityExecutor

			Results   int64 `json:"results"`
			Rated     int64 `json:"rated"`
			Good      int64 `json:"good"`
			Edited    int64 `json:"edited"`
			Discarded int64 `json:"discarded"`
		} `json:"ratings"`
		Judged []struct {
			qualityExecutor

			Judged int64       `json:"judged"`
			Mean   json.Number `json:"mean_score"`
			Best   int64       `json:"best"`
		} `json:"judged"`
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		return "", fmt.Errorf("read quality rows: %w", err)
	}

	var out strings.Builder
	for _, row := range report.Attempts {
		fmt.Fprintf(&out, "%s attempts=%d ok=%d schema=%d failed=%d preempted=%d wall=%.0fs\n",
			row.columns(), row.Attempts, row.Succeeded, row.SchemaViolations, row.Failed, row.Preempted, row.MeanWall)
	}
	for _, row := range report.Ratings {
		fmt.Fprintf(&out, "%s results=%d rated=%d good=%d edited=%d discarded=%d\n",
			row.columns(), row.Results, row.Rated, row.Good, row.Edited, row.Discarded)
	}
	for _, row := range report.Judged {
		fmt.Fprintf(&out, "%-22s judged %-10s %-32s answers=%d mean=%s best=%d\n",
			row.Queue, row.Provider, row.Model, row.Judged, reportNumber(row.Mean), row.Best)
	}
	return out.String(), nil
}

func (executor qualityExecutor) columns() string {
	return fmt.Sprintf("%-22s %-6s %-10s %-32s", executor.Queue, executor.Tier, executor.Provider, executor.Model)
}
