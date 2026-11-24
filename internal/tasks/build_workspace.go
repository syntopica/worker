package tasks

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxWorkspaceBytes is the most input one task may be given: 64 MiB.
const maxWorkspaceBytes = 64 << 20

// buildWorkspace copies each manifest entry to the same relative path of a
// fresh workspace (spec 8: nothing else is visible), and names any doubt.
//
// An entry must resolve inside root after symlinks, so a link planted in the
// input store cannot hand the runner a file from elsewhere, and outside
// denied (the worker's state), which a broad root may contain.
func buildWorkspace(root string, inputs []string, workspace, denied string) reason {
	if len(inputs) > 0 && root == "" {
		return inputsNotAllowed
	}

	total := int64(0)

	for _, entry := range inputs {
		source, size, why := admit(root, entry, denied)
		if why != "" {
			return why
		}

		if total += size; total > maxWorkspaceBytes {
			return inputTooLarge
		}

		if copyFile(source, filepath.Join(workspace, entry)) != nil {
			return nodeError
		}
	}

	return ""
}

// admit resolves one manifest entry to the regular file it names and its
// size, or names why it may not be read.
func admit(root, entry, denied string) (source string, size int64, why reason) {
	source, err := filepath.EvalSymlinks(filepath.Join(root, entry))
	if err != nil || !filepath.IsLocal(entry) || !within(root, source) {
		return "", 0, inputMissing
	}

	info, err := os.Stat(source)
	if err != nil || !info.Mode().IsRegular() {
		return "", 0, inputMissing
	}

	if denied != "" && within(denied, source) {
		return "", 0, inputDenied
	}

	return source, info.Size(), ""
}

// within reports whether path, already resolved, is parent or under it.
func within(parent, path string) bool {
	if resolved, err := filepath.EvalSymlinks(parent); err == nil {
		parent = resolved
	}

	inside, err := filepath.Rel(parent, path)

	return err == nil && inside != ".." && !strings.HasPrefix(inside, ".."+string(filepath.Separator))
}

// copyFile copies source to target, making target's directory.
func copyFile(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), privateDir); err != nil {
		return err //nolint:wrapcheck // the caller reports a fixed code, never the error.
	}

	//nolint:gosec // G304: source was resolved inside the profile's input root.
	from, err := os.Open(source)
	if err != nil {
		return err //nolint:wrapcheck // as above.
	}

	defer func() { _ = from.Close() }()

	//nolint:gosec // G304: target is inside the attempt's own scratch directory.
	to, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, privateFile)
	if err != nil {
		return err //nolint:wrapcheck // as above.
	}

	if _, err = io.Copy(to, from); err != nil {
		_ = to.Close()

		return err //nolint:wrapcheck // as above.
	}

	return to.Close() //nolint:wrapcheck // as above.
}
