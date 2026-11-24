package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestReportOutput(t *testing.T) {
	t.Parallel()
	status := `{"queues":{"work":{"states":{"queued":1},"done_1h":0,"wasted_1h_s":7,"oldest_queued_s":null}},"nodes":{"local":{"reason":"user_active","idle_s":3.2,"resident":["model-a"],"age_s":2,"loads_1h":4,"unloads_1h":3,"last_release":{"code":"user_active","age_s":6}}},"cooldowns":{"codex":8},"openrouter":{"spent_today":5,"key_cap":null,"capped":[]}}`
	costs := `{"rows":[{"day":"2026-11-24","provider":"ollama","queue":"work","attempts":2,"succeeded":1,"tokens_in":3,"tokens_out":4,"cost_usd":0.0012,"wall_s":5}]}`
	audit := `{"rows":[{"time":0.5,"action":"reveal","job_id":"job-1","class":"mail","principal":"admin"}]}`
	quality := `{"attempts":[{"queue":"work","tier":"basic","provider":"ollama","model":"model-a","attempts":2,"succeeded":1,"schema_violations":1,"failed":0,"preempted":0,"mean_wall_s":null}],"ratings":[{"queue":"work","tier":"basic","provider":"ollama","model":"model-a","results":1,"rated":1,"good":0,"edited":1,"discarded":0}],"judged":[{"queue":"work","provider":"ollama","model":"model-a","judged":2,"mean_score":0.75,"best":1}]}`
	for _, tc := range []struct {
		name  string
		body  string
		json  bool
		parts []string
	}{
		{name: "status", body: status, parts: []string{"work", `{"queued": 1}`, "done_1h=0 wasted_1h=7s oldest=None", "cooldown codex", "openrouter today 5/- capped=none"}},
		{name: "nodes", body: status, parts: []string{"local", "user_active", "idle=3.2 resident=['model-a'] age=2s last_release=user_active@6s loads_1h=4 unloads_1h=3"}},
		{name: "quality", body: quality, parts: []string{"attempts=2 ok=1 schema=1 failed=0 preempted=0 wall=0s", "results=1 rated=1 good=0 edited=1 discarded=0", "answers=2 mean=0.75 best=1"}},
		{name: "costs", body: costs, parts: []string{"2026-11-24 ollama", "attempts=2 ok=1 in=3 out=4 cost=$0.0012 wall=5s"}},
		{name: "audit", body: audit, parts: []string{"1970-01-01 00:00:00 reveal mail", "job-1 admin"}},
		{name: "nodes", body: status, json: true, parts: []string{`"queues":`, `"nodes":`}},
		{name: "quality", body: quality, json: true, parts: []string{`"attempts":`, `"ratings":`, `"judged":`}},
		{name: "costs", body: costs, json: true, parts: []string{"[", `"cost_usd": 0.0012`}},
		{name: "audit", body: audit, json: true, parts: []string{"[", `"job_id": "job-1"`}},
	} {
		label := tc.name
		if tc.json {
			label += " JSON"
		}
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer
			if err := renderReport(&output, reportSpec(tc.name), json.RawMessage(tc.body), tc.json); err != nil {
				t.Fatal(err)
			}
			for _, part := range tc.parts {
				if !strings.Contains(output.String(), part) {
					t.Errorf("output %q does not contain %q", output.String(), part)
				}
			}
		})
	}
}
