package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path, text string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestAWorkspaceHoldsOnlyTheManifestAndNeverTheWorkersState(t *testing.T) {
	t.Parallel()

	instance, workspace := t.TempDir(), t.TempDir()
	denied := filepath.Join(instance, "worker", "state")

	write(t, filepath.Join(denied, "tokens", "a.token"), "t")
	write(t, filepath.Join(instance, "notes", "a.md"), "n")
	write(t, filepath.Join(instance, "other.md"), "o")

	if why := buildWorkspace(instance, []string{"notes/a.md"}, workspace, denied); why != "" {
		t.Fatalf("refused: %s", why)
	}

	if copied, _ := os.ReadFile(filepath.Join(workspace, "notes", "a.md")); string(copied) != "n" {
		t.Errorf("copied %q", copied)
	}

	if _, err := os.Stat(filepath.Join(workspace, "other.md")); err == nil {
		t.Error("a file outside the manifest was copied")
	}

	outside := filepath.Join(t.TempDir(), "outside.md")
	write(t, outside, "x")

	if err := os.Symlink(outside, filepath.Join(instance, "link.md")); err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		root   string
		entry  string
		reason reason
	}{
		"the worker's state":    {instance, "worker/state/tokens/a.token", inputDenied},
		"a link out of root":    {instance, "link.md", inputMissing},
		"a parent path":         {filepath.Join(instance, "notes"), "../other.md", inputMissing},
		"an absent file":        {instance, "absent.md", inputMissing},
		"a directory":           {instance, "notes", inputMissing},
		"no input root":         {"", "notes/a.md", inputsNotAllowed},
		"a path through a file": {instance, "other.md/x", inputMissing},
	}

	for name, test := range cases {
		if why := buildWorkspace(test.root, []string{test.entry}, workspace, denied); why != test.reason {
			t.Errorf("%s: %q", name, why)
		}
	}

	blocked := filepath.Join(t.TempDir(), "file")
	write(t, blocked, "")

	if why := buildWorkspace(instance, []string{"notes/a.md"}, blocked, denied); why != nodeError {
		t.Errorf("a workspace that cannot be written: %q", why)
	}
}

func TestAWorkspaceIsBoundedInSize(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	big, err := os.Create(filepath.Join(root, "big"))
	if err != nil {
		t.Fatal(err)
	}

	if err = big.Truncate(maxWorkspaceBytes + 1); err != nil {
		t.Fatal(err)
	}

	_ = big.Close()

	if why := buildWorkspace(root, []string{"big"}, t.TempDir(), ""); why != inputTooLarge {
		t.Errorf("an oversized input: %q", why)
	}
}
