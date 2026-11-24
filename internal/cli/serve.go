package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/coordinator"
)

// Serve runs the coordinator for the selected Syntopica instance.
func Serve(ctx context.Context, diagnostics io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	data := config.InstanceDirectory(os.Getenv("SYNTOPICA_DATA"), cwd)
	if data == "" {
		return ErrNoInstance
	}
	dir, err := config.WorkerDirectory(data)
	if err != nil {
		return err
	}
	server, err := coordinator.Listen(ctx, filepath.Join(dir, "config.json"), filepath.Join(dir, "state"), diagnostics)
	if err != nil {
		return err
	}
	return server.Run(ctx)
}
