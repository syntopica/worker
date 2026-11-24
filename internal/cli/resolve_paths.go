package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/syntopica/worker/internal/config"
)

func resolvePaths() (configPath, state string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("resolve working directory: %w", err)
	}
	data := config.InstanceDirectory(os.Getenv("SYNTOPICA_DATA"), cwd)
	if data == "" {
		return "", "", ErrNoInstance
	}
	dir, err := config.WorkerDirectory(data)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(dir, "config.json"), filepath.Join(dir, "state"), nil
}
