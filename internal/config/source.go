package config

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Source is the live configuration: reloaded when config.json changes on
// disk.
//
// An invalid new file is refused with one fixed line on Log and the last
// good configuration stays in force. The listen address is read once at
// startup: changing it still needs a restart.
type Source struct {
	path   string
	log    io.Writer
	mutex  sync.Mutex
	config *Config
	stamp  stamp
}

type stamp struct {
	modified time.Time
	size     int64
}

// NewSource watches path, starting from the configuration already loaded
// from it.
func NewSource(path string, initial *Config, log io.Writer) *Source {
	source := &Source{path: path, log: log, config: initial}
	source.stamp, _ = source.read()

	return source
}

// Current is the configuration in force now.
func (s *Source) Current() *Config {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now, readable := s.read()
	if !readable || now == s.stamp {
		return s.config
	}

	s.stamp = now

	loaded, err := Load(s.path)
	if err != nil {
		_, _ = fmt.Fprintln(s.log, "worker: config reload refused")

		return s.config
	}

	s.config = loaded

	return s.config
}

func (s *Source) read() (stamp, bool) {
	info, err := os.Stat(s.path)
	if err != nil {
		return stamp{}, false
	}

	return stamp{modified: info.ModTime(), size: info.Size()}, true
}
