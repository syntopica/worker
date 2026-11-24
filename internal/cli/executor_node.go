package cli

import (
	"context"
	"runtime"
	"time"

	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/node"
	"github.com/syntopica/worker/internal/ollama"
)

func (e *executor) node(ctx context.Context) error {
	const commandTimeout = 5 * time.Second
	url := e.policy.CoordinatorURL
	if url == "" {
		url = baseURL(e.cfg)
	}
	run := host.Command(commandTimeout, e.diagnostics)
	loop := &node.Node{
		Name: e.name, Config: e.current, Link: e.link(url),
		Backend: func(url string) ollama.Backend { return &ollama.Server{URL: url, HTTP: e.http, Run: run} },
		HTTP:    e.http, Sample: host.DefaultSampler(&e.policy, runtime.GOOS, e.diagnostics),
		Now: time.Now, Sleep: pause, Log: e.diagnostics,
	}
	return loop.Run(ctx)
}
