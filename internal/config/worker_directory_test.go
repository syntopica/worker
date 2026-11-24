package config

import (
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTheLocalFileOutranksTheTrackedConfig(t *testing.T) {
	t.Parallel()

	data := resolve(t.TempDir())

	if got, err := WorkerDirectory(data); err != nil || got != filepath.Join(data, "worker") {
		t.Errorf("default %q %v", got, err)
	}

	put(t, filepath.Join(data, "syntopica.config.json"), `{"worker": {"path": "tracked"}}`)
	put(t, filepath.Join(data, "syntopica.local.json"), `{"worker": {}}`)

	if got, _ := WorkerDirectory(data); got != filepath.Join(data, "tracked") {
		t.Errorf("tracked %q", got)
	}

	put(t, filepath.Join(data, "syntopica.local.json"), `{"worker": {"path": "/somewhere/else"}}`)

	if got, _ := WorkerDirectory(data); got != "/somewhere/else" {
		t.Errorf("local %q", got)
	}

	put(t, filepath.Join(data, "syntopica.local.json"), `{not json`)

	if _, err := WorkerDirectory(data); err == nil {
		t.Error("a broken file was read")
	}
}
