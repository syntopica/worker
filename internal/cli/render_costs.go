package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

func renderCosts(raw json.RawMessage) (string, error) {
	var rows []struct {
		Day       string      `json:"day"`
		Provider  string      `json:"provider"`
		Queue     string      `json:"queue"`
		Attempts  int64       `json:"attempts"`
		Succeeded int64       `json:"succeeded"`
		TokensIn  json.Number `json:"tokens_in"`
		TokensOut json.Number `json:"tokens_out"`
		Cost      float64     `json:"cost_usd"`
		Wall      float64     `json:"wall_s"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return "", fmt.Errorf("read cost rows: %w", err)
	}

	var out strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&out, "%s %-11s %-22s attempts=%d ok=%d in=%s out=%s cost=$%.4f wall=%.0fs\n",
			row.Day, row.Provider, row.Queue, row.Attempts, row.Succeeded, reportNumber(row.TokensIn), reportNumber(row.TokensOut), row.Cost, row.Wall)
	}
	return out.String(), nil
}
