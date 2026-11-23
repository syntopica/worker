package host

import (
	"fmt"
	"io"
	"strings"
)

// logFailedReaders names the readers that failed, and nothing they read.
func logFailedReaders(log io.Writer, failed []string) {
	if len(failed) > 0 {
		_, _ = fmt.Fprintf(log, "worker: host readers failed: %s\n", strings.Join(failed, ","))
	}
}
