package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTheInstanceIsNamedOrFoundUpward(t *testing.T) {
	t.Parallel()

	home, _ := os.UserHomeDir()

	root := resolve(t.TempDir())
	put(t, filepath.Join(root, "syntopica.config.json"), "{}")
	put(t, filepath.Join(root, "repo", ".git", "HEAD"), "")
	put(t, filepath.Join(root, "plain", "deep", "file"), "")

	cases := []struct{ data, cwd, want string }{
		{"~/data", root, resolve(filepath.Join(home, "data"))},
		{root, "/", root},
		{"", filepath.Join(root, "plain", "deep"), root},
		{"", filepath.Join(root, "repo"), ""},
		{"", "/", ""},
	}

	for _, test := range cases {
		if got := InstanceDirectory(test.data, test.cwd); got != test.want {
			t.Errorf("InstanceDirectory(%q, %q) = %q, want %q", test.data, test.cwd, got, test.want)
		}
	}
}
