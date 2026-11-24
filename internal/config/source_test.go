package config

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func stampFile(t *testing.T, path string, second int64) {
	t.Helper()

	at := time.Unix(second, 0)

	err := os.Chtimes(path, at, at)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAChangedFileIsReloadedAndAnInvalidOneIsRefused(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, baseConfig())
	stampFile(t, path, 1)

	initial, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	var log bytes.Buffer

	source := NewSource(path, initial, &log)
	if source.Current() != initial {
		t.Error("an unchanged file was reloaded")
	}

	grown := baseConfig()
	grown["queues"].(object)["pa.new"] = object{} //nolint:errcheck,forcetypeassert // built just above

	err = os.Rename(writeConfig(t, grown), path)
	if err != nil {
		t.Fatal(err)
	}

	stampFile(t, path, 2)

	if _, added := source.Current().Queues["pa.new"]; !added {
		t.Error("the new queue was not read")
	}

	err = os.WriteFile(path, []byte("{not json"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	stampFile(t, path, 3)

	if len(source.Current().Queues) != 3 || log.String() != "worker: config reload refused\n" {
		t.Errorf("queues %d log %q", len(source.Current().Queues), log.String())
	}

	if os.Remove(path) != nil || len(source.Current().Queues) != 3 {
		t.Error("a missing file dropped the configuration")
	}
}
