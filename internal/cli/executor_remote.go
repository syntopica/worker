package cli

import (
	"context"
	"time"

	"github.com/syntopica/worker/internal/remote"
)

func (e *executor) remote(ctx context.Context) error {
	const beat = 15 * time.Second
	if e.policy.OpenRouterKeyFile == "" {
		return userError("node has no openrouter_key_file")
	}
	key, err := readKey(e.policy.OpenRouterKeyFile)
	if err != nil {
		return err
	}
	provider := &remote.OpenRouter{API: remote.OpenRouterAPI, Key: key, HTTP: e.http}
	slots := make([]*remote.Loop, e.policy.RemoteSlots)
	for i := range slots {
		slots[i] = &remote.Loop{
			Name: e.name, Config: e.current, Link: e.link(baseURL(e.cfg)), Remote: provider,
			Now: time.Now, Sleep: pause, Beat: beat, Log: e.diagnostics,
		}
	}
	return remote.RunSlots(ctx, slots)
}
