package cli

import (
	"os"
	"path/filepath"
	"strings"
)

func readKey(path string) (string, error) {
	const unreadable = userError("the OpenRouter key file cannot be read")
	if relative, ok := strings.CutPrefix(path, "~/"); ok {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", unreadable
		}
		path = filepath.Join(home, relative)
	}
	key, err := os.ReadFile(path)
	if err != nil {
		return "", unreadable
	}
	return strings.TrimSpace(string(key)), nil
}
