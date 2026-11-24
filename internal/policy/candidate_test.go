package policy_test

import (
	"strings"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/policy"
)

const (
	bulk   = "pa.bulk"
	live   = "pa.live"
	modelA = "model-a"
	modelB = "model-b"
	remote = "vendor/model:free"
)

func fixture() *config.Config {
	return &config.Config{
		Models: map[string]config.ModelPin{
			modelA: {Name: modelA, ColdGB: 30, WarmGB: 2},
			modelB: {Name: modelB, ColdGB: 20, WarmGB: 2},
		},
		Nodes: map[string]config.NodePolicy{
			"node-a": {Name: "node-a", Trust: "owner", MemoryBudgetGB: 44},
			"node-g": {Name: "node-g", Trust: "guest", MemoryBudgetGB: 8},
		},
		Queues: map[string]config.QueuePolicy{
			bulk: {Name: bulk, RunWhen: "idle", Weight: 1, MaxModelAge: time.Hour, LocalAfter: 10 * time.Minute},
			live: {Name: live, RunWhen: "active_ok", Weight: 3, MaxModelAge: time.Hour},
		},
		Profiles: map[string]config.TaskProfile{},
		Privacy:  map[string][]string{"mail": {"ollama"}, "public": {"ollama", "openrouter", "runner"}},
		Trust:    map[string][]string{"mail": {"owner"}, "public": {"owner", "guest"}},
	}
}

// routed is the fixture with a runner route, its fallback, an on-demand
// profile and an OpenRouter route on the bulk queue.
func routed() *config.Config {
	cfg := fixture()
	cfg.Profiles = map[string]config.TaskProfile{
		"bulk.agy":   {Name: "bulk.agy", Runner: "agy", Model: "fast", Privacy: []string{"public"}},
		"bulk.codex": {Name: "bulk.codex", Runner: "codex", Privacy: []string{"public"}, Nodes: []string{"node-a"}},
		"max":        {Name: "max", Runner: "max-lane", Privacy: []string{"public"}, OnDemand: true},
	}
	queue := cfg.Queues[bulk]
	queue.Runner = &config.RunnerRoute{Profile: "bulk.agy", After: 100 * time.Second, Fallbacks: []string{"bulk.codex"}}
	queue.OpenRouter = &config.OpenRouterRoute{Models: map[string]string{modelA: remote}, After: 100 * time.Second}
	queue.Fallbacks = map[string][]string{"bulk.agy": {"bulk.codex"}}
	cfg.Queues[bulk] = queue

	return cfg
}

func cand(id string, change ...func(*policy.Candidate)) *policy.Candidate {
	c := &policy.Candidate{JobID: id, Queue: bulk, Model: modelA, Priority: 50, Privacy: "mail", Kind: "inference"}
	for _, apply := range change {
		apply(c)
	}

	return c
}

func public(c *policy.Candidate) { c.Privacy = "public" }

func task(profile string) func(*policy.Candidate) {
	return func(c *policy.Candidate) { c.Kind, c.Model, c.Privacy = "task", profile, "public" }
}

func pinned(pin string) func(*policy.Candidate) {
	return func(c *policy.Candidate) { c.Pin = pin }
}

func request(change ...func(*policy.LeaseRequest)) *policy.LeaseRequest {
	req := &policy.LeaseRequest{Node: "node-a", FreeGB: 44, CurrentIdle: 900}
	for _, apply := range change {
		apply(req)
	}

	return req
}

func resident(model string) func(*policy.LeaseRequest) {
	return func(req *policy.LeaseRequest) { req.ResidentModel = model }
}

func active(req *policy.LeaseRequest) { req.UserActive = true }

func owner() policy.Node { return policy.Node{Name: "node-a", Trust: "owner"} }

func ids(candidates []policy.Candidate) string {
	var out strings.Builder
	for index := range candidates {
		out.WriteString(candidates[index].JobID + "=" + candidates[index].Model + " ")
	}

	return out.String()
}

func list(candidates ...*policy.Candidate) []policy.Candidate {
	out := make([]policy.Candidate, 0, len(candidates))
	for _, c := range candidates {
		out = append(out, *c)
	}

	return out
}
