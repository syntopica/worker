package cli

import "time"

func (c *command) executorArgs() (executorArgs, error) {
	args := executorArgs{}
	synopsis := c.name + " --name NAME"
	if c.name == tasksCommand {
		synopsis += " [--profile PROFILE] [--slots COUNT] [--snapshot] [--override-pins]"
	}
	flags := c.flags(synopsis)
	flags.StringVar(&args.name, "name", "", "node name")
	var snapshot bool
	if c.name == tasksCommand {
		flags.StringVar(&args.profile, "profile", "", "on-demand profile")
		flags.IntVar(&args.slots, "slots", 0, "parallel slots")
		flags.BoolVar(&snapshot, "snapshot", false, "drain only jobs submitted before startup")
		flags.BoolVar(&args.overridePins, "override-pins", false, "allow substituting comparison executors")
	}
	if err := c.parse(flags, 0); err != nil {
		return args, err
	}
	if !args.valid(emptyProfile(flags, args.profile)) {
		flags.Usage()
		return args, errUsage
	}
	if snapshot && args.profile == "" {
		return args, userError("--snapshot requires --profile")
	}
	if args.overridePins && args.profile == "" {
		return args, userError("--override-pins requires --profile")
	}
	if snapshot {
		now := time.Now()
		seconds := float64(now.Unix()) + float64(now.Nanosecond())/float64(time.Second)
		args.createdBefore = &seconds
	}
	return args, nil
}
