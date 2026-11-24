package tasks

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/config"
)

func TestNoRunnerPutsThePromptInItsArguments(t *testing.T) {
	t.Parallel()

	// Each would lift the sandbox or run a subcommand as an argument.
	hostilePrompts := []string{
		"--dangerously-bypass-approvals-and-sandbox -c sandbox_mode=danger-full-access", "resume", "review", "--help",
	}

	for _, name := range []string{"codex", "agy", "claude", "cursor", "max-lane", "another"} {
		for _, prompt := range hostilePrompts {
			call := runnerFor(name).invoke(&job{profile: &config.TaskProfile{Model: "m"}, prompt: prompt, workspace: "/w", scratch: "/s"})

			if slices.Contains(call.argv, prompt) || !strings.Contains(call.stdin, prompt) {
				t.Errorf("%s: prompt %q in argv %q, stdin %q", name, prompt, call.argv, call.stdin)
			}
		}
	}
}

func TestCodexInvocation(t *testing.T) {
	t.Parallel()

	bare := codex().invoke(&job{profile: &config.TaskProfile{}, workspace: "/w", scratch: "/s"})
	want := []string{
		"codex", "exec", "-", "-C", "/w", "-s", "read-only", "--skip-git-repo-check", "--json", "-o", "/s/last-message.txt",
		"--ephemeral", "--ignore-user-config", "--disable", "shell_tool", "--disable", "shell_snapshot",
	}

	if !slices.Equal(bare.argv, want) || bare.answerFile != "/s/last-message.txt" {
		t.Errorf("bare %q", bare.argv)
	}

	full := codex().invoke(&job{
		profile:    &config.TaskProfile{Command: "/bin/codex", Model: "m", Reasoning: "high", InputRoot: "/store"},
		workspace:  "/w",
		scratch:    "/s",
		schemaFile: "/s/schema.json",
	})
	tail := []string{"--ignore-user-config", "-m", "m", "-c", "model_reasoning_effort=high", "--output-schema", "/s/schema.json"}

	if full.argv[0] != "/bin/codex" || !slices.Equal(full.argv[len(full.argv)-len(tail):], tail) {
		t.Errorf("full %q", full.argv)
	}
}

func TestCodexUsageSumsCompletedTurnsAndSkipsNoise(t *testing.T) {
	t.Parallel()

	turns := `{"type": "turn.completed", "usage": {"input_tokens": 100, "output_tokens": 7}}` + "\n" +
		"a warning that is not JSON\n[1]\n" +
		`{"type": "turn.completed", "usage": {"input_tokens": 20, "output_tokens": 5}}` + "\n"

	cases := map[string]struct {
		stdout string
		want   map[string]any
	}{
		"captured turn":    {fixture(t, "codex_events.jsonl"), map[string]any{"tokens_in": int64(24525), "tokens_out": int64(5)}},
		"summed turns":     {turns, map[string]any{"tokens_in": int64(120), "tokens_out": int64(12)}},
		"a count in words": {`{"type": "turn.completed", "usage": {"input_tokens": "many", "output_tokens": 3}}`, map[string]any{"tokens_out": int64(3)}},
		"a failed run":     {fixture(t, "codex_failed_events.jsonl"), map[string]any{}},
		"no events":        {"ok\n", map[string]any{}},
	}

	for name, test := range cases {
		if got := codex().usage(test.stdout); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: %v", name, got)
		}
	}
}

func TestCodexFailuresKeepItsErrorsAndDropWarningsAndAnswers(t *testing.T) {
	t.Parallel()

	text := codex().failures(fixture(t, "codex_failed_events.jsonl"))
	if strings.Count(text, "invalid_request_error") != 2 || strings.Contains(text, "Under-development") {
		t.Errorf("failures %q", text)
	}

	if got := codex().failures(`{"type": "item.completed", "item": {"type": "agent_message", "text": "401"}}`); got != "" {
		t.Errorf("an answer named a failure: %q", got)
	}

	if got := codex().failures(`{"type": "turn.failed", "error": "plain"}` + "\n" + `{"type": "error", "message": 3}`); got != "" {
		t.Errorf("a message that is not text: %q", got)
	}
}
