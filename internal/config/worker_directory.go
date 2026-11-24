package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WorkerDirectory is where an instance keeps worker's configuration and
// state: worker.path from the local file, then the tracked one, else
// "worker".
//
// A relative value resolves against the file that declared it, the same
// rule the other Syntopica engines follow.
func WorkerDirectory(data string) (string, error) {
	for _, name := range []string{"syntopica.local.json", "syntopica.config.json"} {
		content, err := os.ReadFile(filepath.Join(data, name)) //nolint:gosec // the instance's own files
		if err != nil {
			continue
		}

		var file struct {
			Worker struct {
				Path string `json:"path"`
			} `json:"worker"`
		}

		err = json.Unmarshal(content, &file)
		if err != nil {
			return "", fmt.Errorf("%s: %w", name, err)
		}

		if file.Worker.Path == "" {
			continue
		}

		if filepath.IsAbs(file.Worker.Path) {
			return resolve(file.Worker.Path), nil
		}

		return resolve(filepath.Join(data, file.Worker.Path)), nil
	}

	return resolve(filepath.Join(data, "worker")), nil
}
