package cli

import "context"

func (c *command) executor(ctx context.Context) (int, error) {
	arguments, err := c.executorArgs()
	if err != nil {
		return usageExit, err
	}
	runtime, err := loadExecutor(arguments.name, c.errOut)
	if err != nil {
		return usageExit, err
	}
	switch c.name {
	case nodeCommand:
		err = runtime.node(ctx)
	case remoteCommand:
		err = runtime.remote(ctx)
	default:
		err = runtime.tasks(ctx, arguments, c.out)
	}
	if err != nil {
		return usageExit, err
	}
	return 0, nil
}
