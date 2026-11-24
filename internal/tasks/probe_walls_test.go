package tasks

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
)

func TestTheProbeMapsEachConfiguredRunner(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{RunnerQuota: map[string]config.RunnerQuota{
		"cursor": {Provider: "cursor-usage"}, "codex": {Provider: "codex-usage"}, "agy": {Provider: "agy-usage"},
	}}
	usages := map[string]map[string]any{"cursor-usage": {"primary": window(100)}, "agy-usage": {"primary": window(10)}}

	walls := probeWalls(t.Context(), cfg, time.Unix(0, 0), func(_ context.Context, provider string) map[string]any {
		return usages[provider]
	})

	if !reflect.DeepEqual(walls, map[string]float64{"cursor": laterTS}) {
		t.Errorf("walls %v", walls)
	}
}

func TestModelWindowsRestOnlyTheProfilesWhoseModelTheSpentWindowMeters(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Profiles: map[string]config.TaskProfile{
			"p.gemini": {Runner: "runner-a", Model: "Gemini-Pro"},
			"p.claude": {Runner: "runner-a", Model: "claude-sonnet"},
			"p.other":  {Runner: "runner-b", Model: "gemini-pro"},
			"p.bare":   {Runner: "runner-a"},
		},
		RunnerQuota:        map[string]config.RunnerQuota{"runner-a": {Provider: "provider-a"}},
		RunnerModelWindows: map[string]map[string][]string{"runner-a": {"gemini": {"gemini"}, "3p": {"claude", "gpt"}}},
	}
	usage := map[string]any{
		"primary": window(100),
		"extraRateWindows": []any{
			map[string]any{"id": "quota-gemini-weekly", "window": window(98)},
			map[string]any{"id": "quota-3p-weekly", "window": window(0)},
		},
	}

	walls := probeWalls(t.Context(), cfg, time.Unix(0, 0), func(context.Context, string) map[string]any { return usage })

	if !reflect.DeepEqual(walls, map[string]float64{config.CooldownKey("runner-a", "Gemini-Pro"): laterTS}) {
		t.Errorf("walls %v", walls)
	}
}
