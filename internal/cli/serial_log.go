package cli

import (
	"fmt"
	"io"
	"sync"
)

type serialLog struct {
	mu  sync.Mutex
	out io.Writer
}

func (l *serialLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n, err := l.out.Write(p)
	if err != nil {
		return n, fmt.Errorf("write executor diagnostic: %w", err)
	}
	return n, nil
}
