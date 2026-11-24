package cli

import (
	"encoding/json"

	"github.com/syntopica/worker/internal/contract"
)

// runResult reads only what the CLI collects from the result.
type runResult struct {
	ResultID string             `json:"result_id"`
	Control  *string            `json:"control"`
	Detail   json.RawMessage    `json:"detail"`
	Output   json.RawMessage    `json:"output"`
	Executor *contract.Executor `json:"executor"`
}
