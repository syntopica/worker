package jobs_test

import (
	"encoding/json"
	"testing"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/jobs"
	"github.com/syntopica/worker/internal/policy"
)

func TestAdvanceShadowGroupsRecordsTheBlindedVerdictOnce(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	configureShadow(t, data)
	finishAnswer(t, data, "model", 1)
	finishAnswer(t, data, "comparison", 3)
	if count, err := jobs.AdvanceShadowGroups(ctx, data.store, data.cfg, 5); err != nil || count != 1 {
		t.Fatalf("AdvanceShadowGroups() = %d, %v; want one judge queued", count, err)
	}
	finishJudge(t, data)
	if count, err := jobs.AdvanceShadowGroups(ctx, data.store, data.cfg, 8); err != nil || count != 1 {
		t.Fatalf("AdvanceShadowGroups() = %d, %v; want one judge recorded", count, err)
	}
	scores, err := jobs.ReadJudgements(ctx, data.store, 0)
	if err != nil || len(scores) != 2 {
		t.Fatalf("ReadJudgements() = %v, %v; want two executors scored", scores, err)
	}
	if count, advanceErr := jobs.AdvanceShadowGroups(ctx, data.store, data.cfg, 9); advanceErr != nil || count != 0 {
		t.Fatalf("AdvanceShadowGroups(repeat) = %d, %v; want zero", count, advanceErr)
	}
}

func configureShadow(t *testing.T, data jobTestStore) {
	t.Helper()
	queue := data.cfg.Queues["work"]
	queue.Shadow = &config.ShadowPolicy{Rate: 1, Targets: []string{"ollama:comparison"}, Judge: "judge", MaxPending: 10}
	data.cfg.Queues["work"] = queue
	data.cfg.Models["comparison"] = config.ModelPin{ColdGB: 2}
	data.cfg.Privacy["mail"] = []string{"ollama", "runner"}
	data.cfg.Profiles = map[string]config.TaskProfile{"judge": {Runner: "judge-cli", Privacy: []string{"mail"}}}
	submit(t, data.store, data.cfg, document)
}

func finishAnswer(t *testing.T, data jobTestStore, model string, now float64) {
	t.Helper()
	lease := leaseAt(t, localLeaseRequest(), data, now)
	if lease.Model != model {
		t.Fatalf("LeaseJob().Model = %q; want %q", lease.Model, model)
	}
	complete(t, data, lease.AttemptID, lease.Generation, &jobs.CompletionReport{
		Outcome: "succeeded", Output: map[string]any{"text": model + " answer"},
		Executor: map[string]any{"provider": "ollama", "model": model},
	}, now+1)
}

func finishJudge(t *testing.T, data jobTestStore) {
	t.Helper()
	lease := leaseAt(t, &policy.LeaseRequest{Node: "local", Kind: "task"}, data, 6)
	var input map[string]any
	if err := json.Unmarshal(lease.Input, &input); err != nil || input["labels"] == nil || lease.Model != "judge" {
		t.Fatalf("judge lease = %v, %v; want a blinded judge with local labels", lease, err)
	}
	complete(t, data, lease.AttemptID, lease.Generation, &jobs.CompletionReport{
		Outcome: "succeeded",
		Output:  map[string]any{"json": map[string]any{"scores": map[string]any{"A": 5, "B": 4}, "best": "A"}},
	}, 7)
}

func complete(t *testing.T, data jobTestStore, attempt string, generation int64, report *jobs.CompletionReport, now float64) {
	t.Helper()
	state, err := jobs.CompleteAttempt(t.Context(), data.store, data.cfg, attempt, generation, "local", report, now)
	if err != nil || state != "succeeded" {
		t.Fatalf("CompleteAttempt() = %q, %v; want succeeded", state, err)
	}
}
