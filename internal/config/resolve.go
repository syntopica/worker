package config

import "path/filepath"

// resolve is the absolute path with its symbolic links followed, as far as
// the path exists.
func resolve(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}

	followed, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return absolute
	}

	return followed
}
