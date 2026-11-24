package policy_test

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/policy"
)

func TestTheRunnerRouteFallsBackWhileItsProfileRests(t *testing.T) {
	t.Parallel()

	cfg := routed()
	job := cand("a", public)
	guest := policy.Node{Name: "node-g", Trust: "guest"}
	cases := []struct {
		name    string
		job     *policy.Candidate
		cooling map[string]bool
		node    policy.Node
		want    string
	}{
		{"own profile", job, nil, owner(), "bulk.agy"},
		{"runner rests", job, map[string]bool{"agy": true}, owner(), "bulk.codex"},
		{"pinned model rests", job, map[string]bool{"agy:fast": true}, owner(), "bulk.codex"},
		{"another model rests", job, map[string]bool{"agy:slow": true}, owner(), "bulk.agy"},
		{"fallback pinned to another node", job, map[string]bool{"agy": true}, guest, ""},
		{"excluded runner", cand("a", public, func(c *policy.Candidate) { c.ExcludeProviders = []string{"agy"} }), nil, owner(), "bulk.codex"},
		{"class not allowed for runners", cand("a"), nil, owner(), ""},
		{"a runner answered nothing", cand("a", public, func(c *policy.Candidate) { c.Error = "no_output" }), nil, owner(), ""},
		{"no route", cand("a", public, func(c *policy.Candidate) { c.Queue = live }), nil, owner(), ""},
	}

	for _, c := range cases {
		got, found := policy.RunnerRouteProfile(c.job, cfg, c.cooling, c.node)
		if got != c.want || found != (c.want != "") {
			t.Errorf("%s: profile %q, want %q", c.name, got, c.want)
		}
	}
}

func TestTheLocalModelLeavesAYoungJobToARemoteRoute(t *testing.T) {
	t.Parallel()

	cfg := routed()
	resting := map[string]bool{"agy": true, "codex": true}
	capped := map[string]bool{bulk: true}
	cases := []struct {
		name            string
		job             *policy.Candidate
		cooling, capped map[string]bool
		now             float64
		want            bool
	}{
		{"a runner could take it", cand("a", public), nil, nil, 1, true},
		{"old enough for the local model", cand("a", public), nil, nil, 600, false},
		{"only OpenRouter left", cand("a", public), resting, nil, 1, true},
		{"OpenRouter capped today", cand("a", public), resting, capped, 1, false},
		{"OpenRouter excluded", cand("a", public, func(c *policy.Candidate) { c.ExcludeProviders = []string{"openrouter"} }), resting, nil, 1, false},
		{"no remote may take the class", cand("a"), nil, nil, 1, false},
		{"pinned to the local model", cand("a", pinned("ollama:model-a")), nil, nil, 1, false},
		{"pinned elsewhere", cand("a", pinned("runner:bulk.agy")), nil, nil, 9999, true},
		{"removed queue", cand("a", func(c *policy.Candidate) { c.Queue = "removed" }), nil, nil, 1, false},
	}

	for _, c := range cases {
		if got := policy.HeldForRemote(c.job, cfg, c.cooling, c.capped, owner(), c.now); got != c.want {
			t.Errorf("%s: held = %v", c.name, got)
		}
	}
}

func TestTheLadderListsEveryProviderAClassCouldReach(t *testing.T) {
	t.Parallel()

	cfg := routed()
	cases := []struct{ queue, privacy, model, want string }{
		{bulk, "public", modelA, "agy codex ollama openrouter"},
		{bulk, "public", modelB, "agy codex ollama"},
		{bulk, "mail", modelA, "ollama"},
		{live, "public", modelA, "ollama"},
		{bulk, "secret", modelA, ""},
	}

	for _, c := range cases {
		got := slices.Sorted(maps.Keys(policy.LadderProviders(cfg, c.queue, c.privacy, c.model)))
		if want := strings.Fields(c.want); !slices.Equal(got, want) {
			t.Errorf("%s %s %s: providers %v, want %v", c.queue, c.privacy, c.model, got, want)
		}
	}
}

func TestQueuesPastTheirDailyCapAreCapped(t *testing.T) {
	t.Parallel()

	cfg := routed()
	two, five := 2, 5
	cfg.Queues[bulk].OpenRouter.DailyCap = &two

	if capped := policy.CappedQueues(cfg, map[string]int{bulk: 1, live: 9}); len(capped) != 0 {
		t.Errorf("capped below its own cap: %v", capped)
	}

	if capped := policy.CappedQueues(cfg, map[string]int{bulk: 2}); !capped[bulk] {
		t.Error("not capped at its own cap")
	}

	cfg.Queues[bulk].OpenRouter.DailyCap, cfg.Queues[bulk].OpenRouter.DailyKeyCap = nil, &five
	if capped := policy.CappedQueues(cfg, map[string]int{bulk: 1, live: 4}); !capped[bulk] || capped[live] {
		t.Errorf("the key cap counts every queue's spend: %v", capped)
	}
}
