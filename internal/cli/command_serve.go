package cli

import (
	"context"
	"errors"
	"syscall"
)

func (c *command) serve(ctx context.Context) (int, error) {
	if err := c.parse(c.flags("serve"), 0); err != nil {
		return usageExit, err
	}
	err := Serve(ctx, c.errOut)
	if errors.Is(err, syscall.EADDRINUSE) {
		return usageExit, userError("coordinator address is already in use")
	}
	if err != nil {
		return usageExit, err
	}
	return 0, nil
}
