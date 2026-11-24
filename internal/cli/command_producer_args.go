package cli

import "time"

func (c *command) producerArgs() (producerArgs, error) {
	const (
		defaultWait = time.Hour
		defaultPoll = 5 * time.Second
	)
	args := producerArgs{}
	synopsis, count := producerUsage(c.name)
	flags := c.flags(synopsis)
	flags.StringVar(&args.principal, "producer", "", "producer name")
	wait, poll := defaultWait.Seconds(), defaultPoll.Seconds()
	switch c.name {
	case jobsCommand:
		flags.StringVar(&args.queue, "queue", "", "queue name")
		flags.Bool("json", false, "print JSON")
	case runCommand:
		flags.Float64Var(&wait, "wait", wait, "wait seconds")
		flags.Float64Var(&poll, "poll", poll, "poll interval in seconds")
	}
	if err := c.parse(flags, count); err != nil {
		return args, err
	}
	args.values = flags.Args()
	if !args.valid(c.name, wait, poll) {
		flags.Usage()
		return args, errUsage
	}
	args.wait, args.poll = time.Duration(wait*float64(time.Second)), time.Duration(poll*float64(time.Second))
	return args, nil
}
