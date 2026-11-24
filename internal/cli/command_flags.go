package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

func (c *command) flags(synopsis string) *flag.FlagSet {
	flags := flag.NewFlagSet(c.name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	c.synopsis = "usage: worker " + synopsis
	flags.Usage = func() { _, _ = fmt.Fprintln(c.errOut, c.synopsis) }
	return flags
}

func (c *command) parse(flags *flag.FlagSet, positionals int) error {
	// Use the standard parser for values, with an argument order that also
	// accepts flags after positionals, as the original CLI did.
	usage := flags.Usage
	flags.Usage = func() {} // Parse must not choose the help stream.
	err := flags.Parse(flagsFirst(flags, c.args))
	flags.Usage = usage
	switch {
	case errors.Is(err, flag.ErrHelp):
		_, _ = fmt.Fprintln(c.out, c.synopsis)
		return flag.ErrHelp
	case err != nil || flags.NArg() != positionals:
		flags.Usage()
		return errUsage
	default:
		return nil
	}
}
