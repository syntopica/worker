package config

import (
	"os"
	"path/filepath"
	"strings"
)

// InstanceDirectory is the syntopica data directory this invocation serves:
// SYNTOPICA_DATA first, then the instance enclosing cwd, else "".
func InstanceDirectory(data, cwd string) string {
	if data != "" {
		if rest, home := strings.CutPrefix(data, "~/"); home {
			directory, _ := os.UserHomeDir()
			data = filepath.Join(directory, rest)
		}

		return resolve(data)
	}

	return findDataDirectory(cwd)
}

// findDataDirectory walks upward to the nearest ancestor holding
// syntopica.config.json.
//
// The walk stops at a repository boundary: a directory carrying .git that
// is not itself an instance ends the search, because a command run inside
// one repository must not silently read the instance that happens to
// enclose it. The instance is a repository too, so the config check comes
// first.
func findDataDirectory(start string) string {
	for directory := resolve(start); ; directory = filepath.Dir(directory) {
		info, err := os.Stat(filepath.Join(directory, "syntopica.config.json"))
		if err == nil && info.Mode().IsRegular() {
			return directory
		}

		_, err = os.Lstat(filepath.Join(directory, ".git"))
		if err == nil || directory == filepath.Dir(directory) {
			return ""
		}
	}
}
