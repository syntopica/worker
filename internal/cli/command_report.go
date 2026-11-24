package cli

import (
	"context"
	"strconv"
)

func (c *command) report(ctx context.Context) (int, error) {
	const defaultDays = 7.0

	spec := reportSpec(c.name)
	flags := c.flags(c.name + " " + spec.synopsis)
	jsonOutput := flags.Bool("json", false, "print the report as JSON")
	days := defaultDays
	if spec.timed {
		flags.Float64Var(&days, "days", defaultDays, "report window in days")
	}
	if err := c.parse(flags, 0); err != nil {
		return usageExit, err
	}
	admin, err := c.client("admin")
	if err != nil {
		return usageExit, err
	}
	path := spec.path
	if spec.timed {
		path += "?days=" + strconv.FormatFloat(days, 'f', -1, 64)
	}
	raw, err := admin.AdminReport(ctx, path)
	if err != nil {
		return failedExit, err
	}
	if err := renderReport(c.out, spec, raw, *jsonOutput); err != nil {
		return failedExit, err
	}
	return 0, nil
}
