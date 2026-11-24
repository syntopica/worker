package cli

func producerUsage(name string) (synopsis string, positionals int) {
	const rateArguments = 3
	switch name {
	case jobsCommand:
		return "jobs --producer NAME --queue QUEUE [--json]", 0
	case cancelCommand:
		return "cancel --producer NAME JOB_ID", 1
	case rateCommand:
		return "rate --producer NAME JOB_ID RESULT_ID {good,edited,discarded}", rateArguments
	case runCommand:
		return "run --producer NAME [--wait SECONDS] [--poll SECONDS] FILE", 1
	default:
		return "submit --producer NAME FILE", 1
	}
}
