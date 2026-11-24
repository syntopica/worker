package coordinator

import (
	"fmt"
	"io"
	"sync"
)

// diagnosticWriter serializes fixed messages from reload and maintenance.
type diagnosticWriter struct {
	mu   sync.Mutex
	dest io.Writer
}

func (w *diagnosticWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.dest.Write(data)
	if err != nil {
		return n, fmt.Errorf("write coordinator diagnostic: %w", err)
	}
	return n, nil
}
