package policy_test

import (
	"testing"

	"github.com/syntopica/worker/internal/policy"
)

func TestATaskRunsUnderItsProfileOrAQueueFallback(t *testing.T) {
	t.Parallel()

	cfg := routed()
	resting := map[string]bool{"agy": true}
	guest := func(req *policy.LeaseRequest) { req.Node = "node-g" }
	cases := []struct {
		name    string
		job     *policy.Candidate
		req     *policy.LeaseRequest
		cooling map[string]bool
		want    string
	}{
		{"own profile", cand("a", task("bulk.agy")), request(), nil, "bulk.agy"},
		{"fallback while its runner rests", cand("a", task("bulk.agy")), request(), resting, "bulk.codex"},
		{"a pinned task never falls back", cand("a", task("bulk.agy"), pinned("judge")), request(), resting, ""},
		{"fallback not allowed on the node", cand("a", task("bulk.agy")), request(guest), resting, ""},
		{"an on-demand profile is not automatic", cand("a", task("max")), request(), nil, ""},
		{"removed profile", cand("a", task("gone")), request(), nil, ""},
		{"removed queue", cand("a", task("bulk.agy"), func(c *policy.Candidate) { c.Queue = "removed" }), request(), nil, ""},
		{"idle queue under an active user", cand("a", task("bulk.agy")), request(active), nil, ""},
	}

	for _, c := range cases {
		got, found := policy.PickTask(list(c.job), c.req, cfg, nil, c.cooling)
		if got.Model != c.want || found != (c.want != "") {
			t.Errorf("%s: picked %+v", c.name, got)
		}
	}
}

func TestTheTaskLoopLeasesTasksAndRoutedInference(t *testing.T) {
	t.Parallel()

	cfg := routed()
	old := func(c *policy.Candidate) { c.Created = -500 }
	jobs := list(
		cand("task", task("bulk.agy")),
		cand("young", public),
		cand("routed", public, old),
		cand("mail", old),
		cand("unrouted", public, old, func(c *policy.Candidate) { c.Queue = live }),
		cand("member", public, pinned("runner:bulk.codex")),
		cand("judge", task("bulk.agy"), pinned("judge")),
		cand("local", public, old, pinned("ollama:model-a")),
		cand("lane", public, pinned("runner:max")),
		cand("mailpin", pinned("runner:bulk.agy")),
	)

	got := ids(policy.RunnerCandidates(jobs, cfg, nil, owner(), 0))
	if want := "task=bulk.agy routed=bulk.agy member=bulk.codex judge=bulk.agy "; got != want {
		t.Errorf("kept %q, want %q", got, want)
	}

	got = ids(policy.RunnerCandidates(jobs, cfg, map[string]bool{"agy": true}, owner(), 0))
	if want := "task=bulk.agy routed=bulk.codex member=bulk.codex "; got != want {
		t.Errorf("resting: kept %q, want %q", got, want)
	}
}

func TestAnOnDemandProfileTakesWhatItsGrantAllows(t *testing.T) {
	t.Parallel()

	cfg := routed()
	jobs := list(
		cand("inference", public),
		cand("mail"),
		cand("task", task("bulk.agy")),
		cand("judge", task("bulk.agy"), pinned("judge")),
		cand("member", public, pinned("ollama:model-a")),
		cand("original", public, pinned("original")),
		cand("lane", task("max")),
		cand("excluded", public, func(c *policy.Candidate) { c.ExcludeProviders = []string{"max-lane"} }),
		cand("elsewhere", task("bulk.codex")),
	)
	guest := policy.Node{Name: "node-g", Trust: "guest"}

	got := ids(policy.OnDemandCandidates(jobs, cfg, nil, guest, "max", false))
	if want := "inference=max task=max judge=max "; got != want {
		t.Errorf("kept %q, want %q", got, want)
	}

	got = ids(policy.OnDemandCandidates(jobs, cfg, nil, owner(), "max", true))
	if want := "inference=max task=max judge=max member=max elsewhere=max "; got != want {
		t.Errorf("override: kept %q, want %q", got, want)
	}

	for name, cooling := range map[string]map[string]bool{"bulk.agy": nil, "gone": nil, "max": {"max-lane": true}} {
		if kept := policy.OnDemandCandidates(jobs, cfg, cooling, owner(), name, false); kept != nil {
			t.Errorf("%s: an automatic, unknown or resting profile took %v", name, ids(kept))
		}
	}

	pick, found := policy.PickOnDemand(policy.OnDemandCandidates(jobs, cfg, nil, owner(), "max", false), cfg, nil)
	if !found || pick.JobID != "inference" {
		t.Errorf("picked %+v", pick)
	}

	if _, found = policy.PickOnDemand(nil, cfg, nil); found {
		t.Error("picked from nothing")
	}
}
