package cli

import (
	"errors"
	"fmt"
	"slices"

	"github.com/syntopica/worker/internal/auth"
)

func (c *command) token() (int, error) {
	flags := c.flags("token add --kind {producer,node,admin} --name NAME")
	kind := flags.String("kind", "", "principal kind")
	name := flags.String("name", "", "principal name")
	if !c.tokenAction() {
		flags.Usage()
		return usageExit, errUsage
	}
	if err := c.parse(flags, 0); err != nil {
		return usageExit, err
	}
	if *name == "" || !slices.Contains([]string{"producer", nodeCommand, "admin"}, *kind) {
		flags.Usage()
		return usageExit, errUsage
	}
	_, state, err := resolvePaths()
	if err != nil {
		return usageExit, err
	}
	token, err := auth.AddPrincipal(state, *kind, *name)
	if errors.Is(err, auth.ErrInvalidName) {
		return usageExit, userError("invalid principal name")
	}
	if err != nil {
		return usageExit, err
	}
	if _, err := fmt.Fprintln(c.out, token); err != nil {
		return failedExit, fmt.Errorf("write token: %w", err)
	}
	return 0, nil
}

func (c *command) tokenAction() bool {
	if len(c.args) > 0 && c.args[0] == "add" {
		c.args = c.args[1:]
		return true
	}
	return slices.Equal(c.args, []string{"-h"}) || slices.Equal(c.args, []string{"--help"})
}
