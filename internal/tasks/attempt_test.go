package tasks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/link"
)

const (
	agyQuota = "error: Individual quota reached. Please upgrade your subscription to increase your limits. Resets in 2h5m."
	// outOfUsage is the sentence cursor prints when its account is spent.
	outOfUsage = "ActionRequiredError: Increase limits for faster responses You're out of usage. " +
		"Switch to Auto, or ask your admin to increase your limit to continue."
	// codexReports answers with the sandbox flags, the workspace's entries
	// and whether HOME reached the runner.
	codexReports = `flags="$5 $6"; seen=$(ls | tr '\n' ' ')
while [ $# -gt 1 ]; do [ "$1" = "-o" ] && out=$2; shift; done
printf '{"flags": "%s", "seen": "%s", "home": "%s"}' "$flags" "$seen" "${HOME-unset}" > "$out"`
)

// attempted runs one task under a profile and returns its report and log.
func attempted(
	ctx context.Context, profile *config.TaskProfile, input string, coordinator *fakeLink,
) (report *link.Completion, log string) {
	loop, written := testLoop(coordinator, testConfig(profile))
	lease := testLease(testProfile)
	lease.Input = json.RawMessage(input)

	report = (&attempt{
		loop: loop, lease: lease, profile: profile, started: loop.Now(),
		executor: contract.Executor{Node: testNode, Provider: profile.Runner, Model: profile.Model},
	}).run(ctx)

	return report, written.String()
}

// ending is a report as "outcome/error_code".
func endingOf(report *link.Completion) string {
	if report == nil {
		return "fenced"
	}

	if report.ErrorCode == nil {
		return report.Outcome + "/"
	}

	return report.Outcome + "/" + *report.ErrorCode
}

// answerOf is a report's JSON answer, decoded.
func answerOf(t *testing.T, report *link.Completion) map[string]any {
	t.Helper()

	raw, _ := report.Output["json"].(json.RawMessage)

	var answer map[string]any
	if err := json.Unmarshal(raw, &answer); err != nil {
		t.Fatalf("answer %q: %v", raw, err)
	}

	return answer
}

func TestCodexRunsReadOnlyInAWorkspaceHoldingOnlyTheManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write(t, filepath.Join(root, "d", "a.md"), "A")
	write(t, filepath.Join(root, "b.md"), "B")

	profile := fake(t, "codex", codexReports)
	profile.InputRoot, profile.EnvUnset = root, []string{"HOME"}

	report, _ := attempted(t.Context(), profile, `{"prompt": "p", "inputs": ["d/a.md"], "output_schema": {"type": "object"}}`, &fakeLink{})
	if endingOf(report) != "succeeded/" {
		t.Fatalf("ended %s", endingOf(report))
	}

	want := map[string]any{"flags": "-s read-only", "seen": "d ", "home": "unset"}
	if got := answerOf(t, report); !reflect.DeepEqual(got, want) {
		t.Errorf("answer %v", got)
	}
}

func TestATaskThatCannotStartIsRefusedWithItsCode(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	write(t, outside, "x")

	if err := os.Symlink(outside, filepath.Join(root, "link.md")); err != nil {
		t.Fatal(err)
	}

	rooted := fake(t, "codex", "")
	rooted.InputRoot = root

	absent := fake(t, "codex", "")
	absent.Command = filepath.Join(root, "absent")

	cases := map[string]struct {
		profile *config.TaskProfile
		input   string
		want    string
	}{
		"a link out of the root": {rooted, `{"prompt": "p", "inputs": ["link.md"]}`, "failed/input_missing"},
		"inputs without a root":  {fake(t, "codex", ""), `{"prompt": "p", "inputs": ["a.md"]}`, "failed/inputs_not_allowed"},
		"a missing binary":       {absent, `{"prompt": "p"}`, "failed/runner_failed"},
		"an unreadable input":    {fake(t, "codex", ""), `[1]`, "failed/node_error"},
	}

	for name, test := range cases {
		report, _ := attempted(t.Context(), test.profile, test.input, &fakeLink{})
		if endingOf(report) != test.want || report.WallS != 0 {
			t.Errorf("%s: %s wall %v", name, endingOf(report), report.WallS)
		}
	}
}

func TestRunnersEndings(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		runner, body, want string
	}{
		"codex credit wall": {
			"codex", `echo "ERROR: Your workspace is out of credits. Ask your workspace owner to refill in order to continue."; exit 1`,
			"failed/quota_wall",
		},
		"codex without an answer": {"codex", "exit 3", "failed/runner_failed"},
		"codex blank answer":      {"codex", `while [ $# -gt 1 ]; do [ "$1" = "-o" ] && printf ' ' > "$2"; shift; done`, "failed/no_output"},
		"codex failure event": {
			"codex", `echo '{"type": "turn.failed", "error": {"message": "429 Too Many Requests"}}'; exit 1`, "failed/rate_limited",
		},
		"agy silent wall":        {"agy", `echo '{"response": null}'`, "failed/quota_wall"},
		"agy incomplete stream":  {"agy", `echo '{"event": "step_update", "text": "partial"}'`, "failed/no_output"},
		"agy incomplete, exit 1": {"agy", `echo '{"event": "step_update", "text": "partial"}'; exit 1`, "failed/runner_failed"},
		"agy incomplete, exit 2": {"agy", `echo '{"event": "step_update", "text": "partial"}'; exit 2`, "failed/runner_failed"},
		"cursor out of usage":    {"cursor", `printf '%s' "` + outOfUsage + `" >&2; exit 1`, "failed/quota_wall"},
		"cursor quoting its wall": {
			"cursor", `printf '{"is_error": false, "result": "{\\"s\\": \\"%s\\"}"}' "` + outOfUsage + `"`, "succeeded/",
		},
	}

	for name, test := range cases {
		report, _ := attempted(t.Context(), fake(t, test.runner, test.body), `{"prompt": "classify"}`, &fakeLink{})
		if endingOf(report) != test.want {
			t.Errorf("%s: %s", name, endingOf(report))
		}

		if _, resets := report.Usage["wall_resets_in_s"]; resets {
			t.Errorf("%s: a reset nobody named", name)
		}
	}
}

func TestAgyReadsStructuredOutputAndUsageThroughStdin(t *testing.T) {
	t.Parallel()

	body := `stdin=$(cat)
case "$stdin" in *'--help\nclassify"}}') ;; *) exit 9 ;; esac
case " $* " in *" --help "*|*" --dangerously-skip-permissions "*) exit 8 ;; esac
case " $* " in *" --sandbox "*" --json-schema "*) ;; *) exit 7 ;; esac
echo '{"event": "result", "result": {"status": "SUCCESS", "response": "prose", "structured_output": {"ok": true},
 "usage": {"input_tokens": 3, "output_tokens": 2}}}' | tr -d '\n'`

	report, _ := attempted(
		t.Context(), fake(t, "agy", body), `{"prompt": "--help\nclassify", "output_schema": {"type": "object"}}`, &fakeLink{},
	)
	if endingOf(report) != "succeeded/" {
		t.Fatalf("ended %s", endingOf(report))
	}

	if got := answerOf(t, report); !reflect.DeepEqual(got, map[string]any{"ok": true}) {
		t.Errorf("answer %v", got)
	}

	if !reflect.DeepEqual(report.Usage, map[string]any{"tokens_in": int64(3), "tokens_out": int64(2)}) {
		t.Errorf("usage %v", report.Usage)
	}
}

func TestCursorTakesThePromptOnStdinAndTheLastObjectOfItsResult(t *testing.T) {
	t.Parallel()

	body := `printf '{"is_error": false, "result": "narrating {not json} then {\\"echo\\": \\"%s\\"}"}' "$(cat)"`

	// The prompt's credential never reaches the runner.
	secret := "sk-" + strings.Repeat("a", 40)

	prompt := `{"prompt": "big ` + secret + ` prompt. Your code is ٤８2913"}`
	report, _ := attempted(t.Context(), fake(t, "cursor", body), prompt, &fakeLink{})
	if endingOf(report) != "succeeded/" {
		t.Fatalf("ended %s", endingOf(report))
	}

	echo, _ := answerOf(t, report)["echo"].(string)
	if echo != "big [REDACTED] prompt. Your code is [REDACTED]" {
		t.Errorf("echo %q", echo)
	}
}

func TestAWallReportsWhenItResetsAndAnUnmatchedFailureStaysOutOfTheLog(t *testing.T) {
	t.Parallel()

	report, _ := attempted(t.Context(), fake(t, "agy", `printf '%s' "`+agyQuota+`" >&2; exit 1`), `{"prompt": "p"}`, &fakeLink{})
	if endingOf(report) != "failed/quota_wall" || report.Usage["wall_resets_in_s"] != 7500.0 {
		t.Errorf("%s usage %v", endingOf(report), report.Usage)
	}

	report, log := attempted(
		t.Context(), fake(t, "agy", `echo "model gemini-x is not available" >&2; exit 1`), `{"prompt": "p"}`, &fakeLink{},
	)
	if endingOf(report) != "failed/runner_failed" || log != "worker: runner_failed agy:m-1\n" {
		t.Errorf("%s log %q", endingOf(report), log)
	}
}

func TestAttemptReportsOverrunFencingAndShutdown(t *testing.T) {
	t.Parallel()

	overrunning := fake(t, "agy", "sleep 60")
	overrunning.Timeout = 200 * time.Millisecond

	cases := map[string]struct {
		profile     *config.TaskProfile
		coordinator *fakeLink
		want        string
		cancel      bool
	}{
		"overrun": {overrunning, &fakeLink{}, "failed/timeout", false},
		"fenced":  {fake(t, "agy", "sleep 60"), &fakeLink{fenced: true}, "fenced", false},
		"stopped": {fake(t, "codex", "sleep 60"), &fakeLink{}, "preempted/node_shutdown", true},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, stop := context.WithCancel(t.Context())
			defer stop()
			if test.cancel {
				timer := time.AfterFunc(100*time.Millisecond, stop)
				defer timer.Stop()
			}
			started := time.Now()

			report, _ := attempted(ctx, test.profile, `{"prompt":"fixture prompt"}`, test.coordinator)
			if endingOf(report) != test.want || time.Since(started) > 10*time.Second {
				t.Errorf("%s after %s, want %s", endingOf(report), time.Since(started), test.want)
			}
		})
	}
}

func TestAFencedTaskIsLoggedAndNeverCompleted(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{fenced: true, leases: []*contract.Lease{testLease(testProfile)}}
	loop, log := testLoop(coordinator, testConfig(fake(t, "codex", "sleep 60")))

	if resting, err := loop.step(t.Context(), loop.Config()); resting || err != nil {
		t.Fatalf("resting %v, err %v", resting, err)
	}

	if len(coordinator.completed) != 0 || log.String() != "worker: task fenced job=j\n" {
		t.Errorf("completed %d, log %q", len(coordinator.completed), log.String())
	}
}

func TestOnDemandDrainsItsProfileThenFlushes(t *testing.T) {
	t.Parallel()

	before := 12.5
	coordinator := &fakeLink{leases: []*contract.Lease{testLease(testProfile), testLease(testProfile), testLease(testProfile)}}
	loop, _ := testLoop(coordinator, testConfig(fake(t, "codex", codexAnswers)))
	loop.Sample = nil

	drain := func() *OnDemand {
		return &OnDemand{Loop: loop, Profile: testProfile, OverridePins: true, CreatedBefore: &before}
	}

	if ran := RunOnDemand(t.Context(), []*OnDemand{drain(), drain()}); ran != 3 {
		t.Errorf("ran %d", ran)
	}

	if got := coordinator.ended(); got != strings.Repeat(" succeeded/", 3)[1:] || coordinator.flushed != 2 {
		t.Errorf("ended %q, flushed %d", got, coordinator.flushed)
	}

	request := coordinator.requests[0]
	if request.Profile != testProfile || !request.OverridePins || request.CreatedBefore != &before || request.UserActive {
		t.Errorf("request %+v", request)
	}
}
