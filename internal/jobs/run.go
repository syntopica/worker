package jobs

import (
	"context"
	"fmt"

	"github.com/syntopica/worker/internal/store"
)

// run executes every statement of a script with the same named parameters.
func run(ctx context.Context, c store.Conn, name string, p params) error {
	if _, err := c.ExecContext(ctx, script(name), p.list()...); err != nil {
		return fmt.Errorf("jobs: %s: %w", name, err)
	}

	return nil
}
