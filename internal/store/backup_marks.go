package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync"
	"time"
)

// backupMarks keeps the payload file out of Time Machine (spec 9: content is
// never backed up) and remembers which file identities it already marked.
type backupMarks struct {
	lock sync.Mutex
	done map[string]os.FileInfo
	// run sets the attribute on one path; tests replace it.
	run func(ctx context.Context, path string) error
	log func(line string)
}

const (
	backupAttribute = "com.apple.metadata:com_apple_backup_excludeItem"
	// backupValue is the binary property list of "com.apple.backupd", in hex:
	// what `tmutil addexclusion` writes.
	backupValue = "62706c69737430305f1011636f6d2e6170706c652e6261636b75706408" +
		"000000000000010100000000000000010000000000000000000000000000001c"
	backupTimeout = 5 * time.Second
)

func newBackupMarks() *backupMarks {
	return &backupMarks{
		done: map[string]os.FileInfo{},
		run:  setBackupAttribute,
		log:  func(line string) { fmt.Fprintln(os.Stderr, line) },
	}
}

// exclude marks the payload file and its sidecars; the metadata file stays
// eligible, since it is the file that is backed up. Only a success is
// remembered, so a failure is retried on the next call and a recreated file
// is marked again. A failure logs its class on one line and is never returned.
func (m *backupMarks) exclude(ctx context.Context, dir string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, name := range []string{payloadFile, payloadFile + "-wal", payloadFile + "-shm"} {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if seen, ok := m.done[path]; ok && os.SameFile(seen, info) {
			continue
		}
		if err := m.run(ctx, path); err != nil {
			m.log("worker: backup exclusion failed: " + errorClass(err))
			continue
		}
		m.done[path] = info
	}
}

func setBackupAttribute(ctx context.Context, path string) error {
	ctx, cancel := context.WithTimeout(ctx, backupTimeout)
	defer cancel()
	//nolint:gosec // fixed binary and attribute; the path is the store's own file
	err := exec.CommandContext(ctx, "/usr/bin/xattr", "-wx", backupAttribute, backupValue, path).Run()
	if err != nil {
		return fmt.Errorf("xattr: %w", err)
	}
	return nil
}

// errorClass names an error without its message, which may hold a path.
func errorClass(err error) string {
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		return fmt.Sprintf("exit %d", exit.ExitCode())
	}
	return reflect.TypeOf(err).String()
}
