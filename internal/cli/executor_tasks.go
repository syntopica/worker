package cli

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/tasks"
)

func (e *executor) tasks(ctx context.Context, args executorArgs, out io.Writer) error {
	count := args.slots
	if count == 0 {
		count = e.policy.TaskSlots
	}
	if args.profile != "" {
		return e.drainTasks(ctx, args, count, out)
	}
	slots := make([]*tasks.Loop, count)
	for i := range slots {
		slots[i] = e.taskSlot(e.current)
		slots[i].Sample = host.DefaultSampler(&e.policy, runtime.GOOS, e.diagnostics)
	}
	slots[0].Usage = (tasks.CodexBar{}).Usage
	return tasks.RunSlots(ctx, slots)
}

func (e *executor) drainTasks(ctx context.Context, args executorArgs, count int, out io.Writer) error {
	slots := make([]*tasks.OnDemand, count)
	for i := range slots {
		slots[i] = &tasks.OnDemand{
			Loop:    e.taskSlot(func() *config.Config { return e.cfg }),
			Profile: args.profile, OverridePins: args.overridePins, CreatedBefore: args.createdBefore,
		}
	}
	ran := tasks.RunOnDemand(ctx, slots)
	if _, err := fmt.Fprintf(out, "worker: on-demand %s ran %d jobs\n", args.profile, ran); err != nil {
		return fmt.Errorf("write on-demand summary: %w", err)
	}
	return nil
}

func (e *executor) taskSlot(current func() *config.Config) *tasks.Loop {
	const beat = 20 * time.Second
	return &tasks.Loop{
		Name: e.name, Config: current, Link: e.link(baseURL(e.cfg)),
		Now: time.Now, Sleep: pause, Beat: beat, Log: e.diagnostics,
	}
}
