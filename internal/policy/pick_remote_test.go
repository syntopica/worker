package policy_test

import (
	"testing"

	"github.com/syntopica/worker/internal/policy"
)

func TestOpenRouterTakesARoutedJobThatWaitedLongEnough(t *testing.T) {
	t.Parallel()

	cfg := routed()
	always := func(*policy.Candidate) bool { return true }
	never := func(*policy.Candidate) bool { return false }
	other := func(c *policy.Candidate) { c.Model = modelB }
	onLive := func(c *policy.Candidate) { c.Queue = live }
	cases := []struct {
		name        string
		job         *policy.Candidate
		runnerFirst func(*policy.Candidate) bool
		now         float64
		want        string
	}{
		{"waits while a runner would take it", cand("a"), always, 99, ""},
		{"escalates after the wait", cand("a"), always, 100, remote},
		{"escalates at once without a runner", cand("a"), never, 1, remote},
		{"no callback means no runner", cand("a"), nil, 1, remote},
		{"unmapped model", cand("a", other), nil, 999, ""},
		{"queue without a route", cand("a", onLive), nil, 999, ""},
		{"pinned to an OpenRouter model", cand("a", onLive, pinned("openrouter:vendor/other")), always, 1, "vendor/other"},
		{"pinned elsewhere", cand("a", pinned("ollama:model-a")), nil, 999, ""},
	}

	for _, c := range cases {
		got, found := policy.PickRemote(list(c.job), cfg, nil, c.now, c.runnerFirst)
		if got.Model != c.want || found != (c.want != "") || (found && got.JobID != "a") {
			t.Errorf("%s: picked %+v", c.name, got)
		}
	}

	urgent := cand("urgent", func(c *policy.Candidate) { c.Priority = 99 })

	got, _ := policy.PickRemote(list(cand("a"), urgent), cfg, nil, 500, nil)
	if got.JobID != "urgent" {
		t.Errorf("priority ignored: %+v", got)
	}

	if !policy.OpenRouterMayTake(cand("a", public), cfg, "guest") || policy.OpenRouterMayTake(cand("a"), cfg, "owner") {
		t.Error("the route ignored the class")
	}
}
