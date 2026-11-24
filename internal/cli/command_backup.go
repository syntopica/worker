package cli

import (
	"context"
	"errors"
	"io/fs"

	"github.com/syntopica/worker/internal/store"
)

func (c *command) backup(ctx context.Context) (int, error) {
	flags := c.flags("backup DEST")
	if err := c.parse(flags, 1); err != nil {
		return usageExit, err
	}
	_, state, err := resolvePaths()
	if err != nil {
		return usageExit, err
	}
	err = store.Backup(ctx, state, flags.Arg(0))
	switch {
	case err == nil:
		return 0, nil
	case errors.Is(err, store.ErrBackupSourceMissing):
		return usageExit, userError("no metadata store")
	case errors.Is(err, store.ErrBackupDestination) && errors.Is(err, fs.ErrExist):
		return usageExit, userError("backup destination already exists")
	case errors.Is(err, store.ErrBackupDestination):
		return usageExit, userError("backup destination cannot be created")
	default:
		return failedExit, err
	}
}
