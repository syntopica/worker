// Command worker runs the worker engine.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/syntopica/worker/internal/cli"
)

func main() {
	os.Exit(run())
}

func run() int {
	args := os.Args[1:]
	ctx := context.Background()
	if len(args) == 0 || !cli.DrainsOnSignal(args[0]) {
		return cli.Run(ctx, args, os.Stdin, os.Stdout, os.Stderr)
	}
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return cli.Run(ctx, args, os.Stdin, os.Stdout, os.Stderr)
}
