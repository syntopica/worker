package policy_test

import (
	"testing"

	"github.com/syntopica/worker/internal/policy"
)

func TestUserActivityAdmitsOnlyActiveOKQueues(t *testing.T) {
	t.Parallel()

	cfg := fixture()
	onLive := func(c *policy.Candidate) { c.Queue = live }

	if policy.CandidateEligible(cand("a"), request(active), cfg) {
		t.Error("an idle queue ran under an active user")
	}

	if !policy.CandidateEligible(cand("b", onLive), request(active), cfg) {
		t.Error("an active_ok queue was held back")
	}

	gone := func(c *policy.Candidate) { c.Queue = "removed" }
	unknown := func(c *policy.Candidate) { c.Model = "model-x" }

	if policy.CandidateEligible(cand("c", gone), request(), cfg) || policy.CandidateEligible(cand("d", unknown), request(), cfg) {
		t.Error("a removed queue or model ran")
	}
}

func TestAResidentModelNeedsWarmMemoryOnly(t *testing.T) {
	t.Parallel()

	cfg := fixture()
	tight := func(req *policy.LeaseRequest) { req.FreeGB = 3 }
	guest := func(req *policy.LeaseRequest) { req.Node = "node-g" }

	if !policy.CandidateEligible(cand("a"), request(resident(modelA), tight), cfg) {
		t.Error("the resident model did not fit its warm cost")
	}
	// Another resident model is evicted, so the cold footprint is checked
	// against the node's whole budget.
	if !policy.CandidateEligible(cand("a"), request(resident(modelB), tight), cfg) {
		t.Error("a cold model did not fit the whole budget")
	}

	if policy.CandidateEligible(cand("a"), request(guest), cfg) {
		t.Error("30 GB cold fit a budget of 8")
	}
}

func TestAParkedJobWaitsForALongEnoughCurrentIdle(t *testing.T) {
	t.Parallel()

	cfg := fixture()
	parked := cand("a", func(c *policy.Candidate) { c.Parked, c.ParkedMinIdle = true, 1200 })
	idle := func(req *policy.LeaseRequest) { req.CurrentIdle = 1300 }

	if policy.CandidateEligible(parked, request(), cfg) || !policy.CandidateEligible(parked, request(idle), cfg) {
		t.Error("parking ignored the current idle")
	}
}

func TestPickJobOrdering(t *testing.T) {
	t.Parallel()

	cfg := fixture()
	at := func(created float64) func(*policy.Candidate) {
		return func(c *policy.Candidate) { c.Created = created }
	}
	other := func(priority int) func(*policy.Candidate) {
		return func(c *policy.Candidate) { c.Model, c.Priority = modelB, priority }
	}
	onLive := func(c *policy.Candidate) { c.Queue = live }
	urgent := func(c *policy.Candidate) { c.Priority = 99 }
	shares := map[string]int{bulk: 2, live: 3}

	cases := []struct {
		name   string
		jobs   []policy.Candidate
		req    *policy.LeaseRequest
		recent map[string]int
		now    float64
		want   string
	}{
		{"resident wins", list(cand("b", other(90), at(1000)), cand("a", at(1000))), request(resident(modelA)), nil, 1100, "a"},
		{"until another model is too old", list(cand("b", other(90), at(1000)), cand("a", at(1000))), request(resident(modelA)), nil, 4601, "b"},
		{"overdue beats higher priority", list(cand("a", at(1000)), cand("b", other(10), at(1000))), request(resident(modelA)), nil, 4601, "b"},
		// pa.live weight 3 with 3 recent runs (1.0) against pa.bulk weight 1 with 2 (2.0).
		{"weighted share", list(cand("old", at(1)), cand("live", onLive, at(5))), request(), shares, 10, "live"},
		{"priority first", list(cand("old", at(1)), cand("live", onLive, at(5)), cand("urgent", urgent, at(9))), request(), nil, 10, "urgent"},
		{"age last", list(cand("young", at(5)), cand("old", at(1)), cand("twin", at(1))), request(), nil, 10, "old"},
	}

	for _, c := range cases {
		got, found := policy.PickJob(c.jobs, c.req, cfg, c.recent, c.now)
		if !found || got.JobID != c.want {
			t.Errorf("%s: picked %q, want %q", c.name, got.JobID, c.want)
		}
	}

	if _, found := policy.PickJob(list(cand("a")), request(active), cfg, nil, 1); found {
		t.Error("an ineligible job was picked")
	}
}
