package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// protectFiles sets the directory to 0700 and each database file and its
// sidecars to 0600, when present.
func protectFiles(dir string) error {
	if err := os.Chmod(dir, privateDir); err != nil {
		return fmt.Errorf("protect state directory: %w", err)
	}
	for _, name := range []string{metaFile, payloadFile} {
		for _, suffix := range []string{"", "-wal", "-shm", "-journal"} {
			err := os.Chmod(filepath.Join(dir, name+suffix), privateFile)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("protect %s: %w", name+suffix, err)
			}
		}
	}
	return nil
}
