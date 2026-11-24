// Command worker runs the worker engine.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/syntopica/worker/internal/cli"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	const usage = "usage: worker serve"
	if wantsHelp(args) {
		_, _ = fmt.Fprintln(os.Stdout, usage)
		return 0
	}
	if len(args) != 1 || args[0] != "serve" {
		_, _ = fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := cli.Serve(ctx, os.Stderr); err != nil {
		switch {
		case errors.Is(err, cli.ErrNoInstance):
			_, _ = fmt.Fprintln(os.Stderr, "worker:", cli.ErrNoInstance)
		case errors.Is(err, syscall.EADDRINUSE):
			_, _ = fmt.Fprintln(os.Stderr, "worker: coordinator address is already in use")
		default:
			_, _ = fmt.Fprintln(os.Stderr, "worker: coordinator failed")
		}
		return 2
	}
	return 0
}

func wantsHelp(args []string) bool {
	if len(args) > 0 && args[0] == "serve" {
		args = args[1:]
	}
	return slices.Equal(args, []string{"-h"}) || slices.Equal(args, []string{"--help"})
}
