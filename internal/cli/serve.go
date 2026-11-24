package cli

import (
	"context"
	"io"

	"github.com/syntopica/worker/internal/coordinator"
)

// Serve runs the coordinator for the selected Syntopica instance.
func Serve(ctx context.Context, diagnostics io.Writer) error {
	path, state, err := resolvePaths()
	if err != nil {
		return err
	}
	server, err := coordinator.Listen(ctx, path, state, diagnostics)
	if err != nil {
		return err
	}
	return server.Run(ctx)
}
