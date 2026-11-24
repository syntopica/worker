package cli

import (
	"context"
	"fmt"
	"io"
)

// Run executes one worker invocation and returns its process exit status.
func Run(ctx context.Context, args []string, in io.Reader, out, diagnostics io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		_, _ = fmt.Fprintln(out, usage)
		return 0
	}
	if len(args) == 0 {
		_, _ = fmt.Fprintln(diagnostics, usage)
		return usageExit
	}
	c := &command{name: args[0], args: args[1:], in: in, out: out, errOut: diagnostics}
	code, err := c.dispatch(ctx)
	return c.outcome(code, err)
}
