package tasks

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"
)

// codexBarWait bounds one usage read.
const codexBarWait = 90 * time.Second

// CodexBar reads providers' usage reports from the CodexBar CLI.
type CodexBar struct {
	// Command is the CodexBar executable; empty means "codexbar" on PATH.
	Command string
}

// Usage is one provider's usage object, or nil when CodexBar is missing,
// slow or unreadable.
func (c CodexBar) Usage(ctx context.Context, provider string) map[string]any {
	command := c.Command
	if command == "" {
		command = "codexbar"
	}

	ctx, cancel := context.WithTimeout(ctx, codexBarWait)
	defer cancel()

	//nolint:gosec // G204: the configured command and a provider name from the configuration.
	out, _ := exec.CommandContext(ctx, command, "usage", "--provider", provider, "--format", jsonFormat).Output()

	var reports []struct {
		Usage map[string]any `json:"usage"`
	}

	if json.Unmarshal(out, &reports) != nil || len(reports) == 0 {
		return nil
	}

	return reports[0].Usage
}
