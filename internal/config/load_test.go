package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

type object = map[string]any

func baseConfig() object {
	return object{
		"listen": "127.0.0.1:8765",
		"models": object{"model-a": object{"num_ctx": 40960, "keep_alive": "5m", "cold_gb": 30, "warm_gb": 2}},
		"queues": object{
			"pa.bulk": object{"run_when": "idle"},
			"pa.live": object{"run_when": "active_ok", "weight": 3},
		},
		"nodes": object{"node-a": object{
			"trust": "owner", "memory_budget_gb": 44,
			"ollama_url": "http://127.0.0.1:11434", "ollama_launchd_label": "label-a",
		}},
		"producers": object{"pa": []string{"pa.bulk", "pa.live"}},
	}
}

func writeConfig(t *testing.T, data object) string {
	t.Helper()

	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "config.json")

	err = os.WriteFile(path, encoded, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

func with(key string, value any) object {
	data := baseConfig()
	data[key] = value

	return data
}

func queueWith(queue object) object {
	data := baseConfig()
	data["queues"] = object{"q": queue}
	data["profiles"] = object{
		"p.one":    object{"runner": "codex", "input_root": "in"},
		"p.two":    object{"runner": "cursor", "input_root": "in"},
		"p.narrow": object{"runner": "cursor", "privacy": []string{"public"}, "input_root": "in"},
		"p.other":  object{"runner": "cursor", "input_root": "elsewhere"},
	}

	return data
}

func TestDefaultsFillEveryOptionalValue(t *testing.T) {
	t.Parallel()

	config, err := Load(writeConfig(t, baseConfig()))
	if err != nil {
		t.Fatal(err)
	}

	bulk := QueuePolicy{
		Name: "pa.bulk", RunWhen: "idle", Weight: 1, MaxOutstanding: 200, RetentionDays: 7, UnackedTTLHours: 72,
		ParkedMinIdle: 10 * time.Minute, MaxModelAge: time.Hour, Tiers: map[string]TierRoute{},
	}
	node := NodePolicy{
		Name: "node-a", Trust: "owner", IdleThreshold: 5 * time.Minute, MemoryBudgetGB: 44,
		OllamaURL: "http://127.0.0.1:11434", OllamaLaunchdLabel: "label-a", MinFreePct: 15, RemoteSlots: 1, TaskSlots: 1,
	}
	limits := [...]int{config.ListenPort, config.MaxPayloadBytes, config.MaxSplit, config.MaxParkedRuns, config.SplitAfterPreemptions}

	if !reflect.DeepEqual(config.Queues["pa.bulk"], bulk) || !reflect.DeepEqual(config.Nodes["node-a"], node) {
		t.Errorf("queue %+v node %+v", config.Queues["pa.bulk"], config.Nodes["node-a"])
	}

	if limits != [...]int{8765, 1_048_576, 8, 3, 3} || config.ListenHost != "127.0.0.1" {
		t.Errorf("limits %v", limits)
	}

	local := []string{"ollama", "local-cpu"}
	if config.Queues["pa.live"].Weight != 3 || !reflect.DeepEqual(config.Privacy["secret"], local) ||
		!reflect.DeepEqual(config.Trust["mail"], []string{"owner"}) {
		t.Errorf("config %+v", config)
	}
}

func TestAnOverrideReplacesOneClassOnly(t *testing.T) {
	t.Parallel()

	override := object{"mail": object{"executors": []string{"ollama"}, "trust": []string{"owner"}}}

	config, err := Load(writeConfig(t, with("privacy", override)))
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(config.Privacy["mail"], []string{"ollama"}) || len(config.Privacy["secret"]) != 2 {
		t.Errorf("privacy %v", config.Privacy)
	}
}

func TestEveryRouteIsRead(t *testing.T) {
	t.Parallel()

	home, _ := os.UserHomeDir()

	data := queueWith(object{
		"profiles":   []string{"p.one", "p.two"},
		"fallbacks":  object{"p.one": []string{"p.two"}},
		"openrouter": object{"models": object{"model-a": "vendor/model:free"}, "after_s": 100, "daily_cap": 5},
		"tiers":      object{"strong": object{"models": []string{"model-a"}, "profiles": object{"p.one": "p.two"}}},
		"runner":     object{"profile": "p.one", "after_s": 5, "fallbacks": []string{"p.two"}},
		"shadow": object{
			"rate": 0.5, "judge": "p.one",
			"targets": []string{"ollama:model-a", "runner:p.two", "openrouter:" + FreeRouterModel},
		},
		"local_after_s": 60,
	})
	data["profiles"].(object)["p.claude"] = object{ //nolint:errcheck,forcetypeassert // built just above
		"runner": "claude", "model": "m", "nodes": []string{"node-a"}, "env": object{"B": "~/b", "A": "a"},
	}
	data["runners"] = object{"codex": object{
		"cooldown_s": 60, "quota_provider": "p", "quota_windows": []string{"w"}, "max_concurrent": 2,
		"model_windows": object{"gemini": []string{"gemini"}},
	}}

	path := writeConfig(t, data)

	config, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	limit, root := 5, resolve(filepath.Join(filepath.Dir(path), "in"))
	queue := QueuePolicy{
		Name: "q", RunWhen: "idle", Weight: 1, MaxOutstanding: 200, RetentionDays: 7, UnackedTTLHours: 72,
		ParkedMinIdle: 10 * time.Minute, MaxModelAge: time.Hour, LocalAfter: time.Minute,
		Profiles:  []string{"p.one", "p.two"},
		Fallbacks: map[string][]string{"p.one": {"p.two"}},
		Tiers:     map[string]TierRoute{"strong": {Models: []string{"model-a"}, Profiles: map[string]string{"p.one": "p.two"}}},
		Runner:    &RunnerRoute{Profile: "p.one", After: 5 * time.Second, Fallbacks: []string{"p.two"}},
		OpenRouter: &OpenRouterRoute{
			Models: map[string]string{"model-a": "vendor/model:free"}, After: 100 * time.Second, ZDR: true, DailyCap: &limit,
		},
		Shadow: &ShadowPolicy{
			Rate: 0.5, Judge: "p.one", MaxPending: 20,
			Targets: []string{"ollama:model-a", "runner:p.two", "openrouter:" + FreeRouterModel},
		},
	}
	claude := TaskProfile{
		Name: "p.claude", Runner: "claude", Model: "m", Privacy: []string{"public", "internal"},
		Timeout: 15 * time.Minute, Nodes: []string{"node-a"}, DeniedRoot: resolve(filepath.Join(filepath.Dir(path), "state")),
		Env: [][2]string{{"A", "a"}, {"B", filepath.Join(home, "b")}},
	}

	if !reflect.DeepEqual(config.Queues["q"], queue) || !reflect.DeepEqual(config.Profiles["p.claude"], claude) {
		t.Errorf("queue %+v profile %+v", config.Queues["q"], config.Profiles["p.claude"])
	}

	runners := Config{
		RunnerCooldown:      map[string]time.Duration{"codex": time.Minute},
		RunnerQuota:         map[string]RunnerQuota{"codex": {Provider: "p", Windows: []string{"w"}}},
		RunnerMaxConcurrent: map[string]int{"codex": 2},
		RunnerModelWindows:  map[string]map[string][]string{"codex": {"gemini": {"gemini"}}},
	}

	if config.Profiles["p.one"].InputRoot != root || !reflect.DeepEqual(config.RunnerCooldown, runners.RunnerCooldown) ||
		!reflect.DeepEqual(config.RunnerQuota, runners.RunnerQuota) ||
		!reflect.DeepEqual(config.RunnerMaxConcurrent, runners.RunnerMaxConcurrent) ||
		!reflect.DeepEqual(config.RunnerModelWindows, runners.RunnerModelWindows) {
		t.Errorf("runners %+v", config)
	}
}

func TestAnInvalidConfigIsRefused(t *testing.T) {
	t.Parallel()

	granted := []string{"p.one", "p.two", "p.narrow", "p.other"}
	profile := func(entry object) object { return with("profiles", object{"p": entry}) }
	route := func(entry object) object { return queueWith(object{"openrouter": entry}) }
	shadow := func(rate float64, judge string, targets ...string) object {
		return queueWith(object{"shadow": object{"rate": rate, "judge": judge, "targets": targets}})
	}

	cases := map[string]struct {
		data object
		want string
	}{
		"unknown run_when":     {queueWith(object{"run_when": "sometimes"}), "run_when"},
		"no retention":         {queueWith(object{"retention_days": 0}), "retention_days"},
		"short retention":      {queueWith(object{"retention_days": 2, "unacked_ttl_hours": 72}), "retention_days"},
		"wildcard listen":      {with("listen", "0.0.0.0:8765"), "loopback"},
		"wildcard v6 listen":   {with("listen", "[::]:8765"), "loopback"},
		"routable listen":      {with("listen", "192.0.2.10:8765"), "loopback"},
		"named listen":         {with("listen", "example.test:8765"), "loopback"},
		"portless listen":      {with("listen", "localhost"), "listen"},
		"wordy port":           {with("listen", "localhost:http"), "listen"},
		"missing section":      {object{"models": object{}}, "required"},
		"windows as list":      {with("runners", object{"r": object{"model_windows": []string{"g"}}}), "model_windows"},
		"windows as string":    {with("runners", object{"r": object{"model_windows": object{"g": "g"}}}), "model_windows"},
		"empty window":         {with("runners", object{"r": object{"model_windows": object{"g": []string{}}}}), "model_windows"},
		"unknown runner":       {profile(object{"runner": "other"}), "runner must be"},
		"agy with inputs":      {profile(object{"runner": "agy", "input_root": "in"}), "input_root"},
		"claude with inputs":   {profile(object{"runner": "claude", "model": "m", "input_root": "in"}), "input_root"},
		"claude unpinned":      {profile(object{"runner": "claude"}), "name a model"},
		"max-lane routed":      {profile(object{"runner": "max-lane", "model": "m"}), "on_demand"},
		"unnamed variable":     {profile(object{"runner": "codex", "env": object{"": "v"}}), "env"},
		"route without models": {route(object{"models": object{}}), "openrouter.models"},
		"paid route":           {route(object{"models": object{"model-a": "vendor/model"}}), "openrouter.models"},
		"paid fallback": {
			route(object{"models": object{"model-a": "v/m:free"}, "fallbacks": []string{"v/m"}}), "fallbacks",
		},
		"zero cap":           {route(object{"models": object{"model-a": "v/m:free"}, "daily_key_cap": 0}), "daily caps"},
		"unknown tier":       {queueWith(object{"tiers": object{"best": object{}}}), "tiers"},
		"unknown tier model": {queueWith(object{"tiers": object{"basic": object{"models": []string{"m"}}}}), "unknown models"},
		"ungranted stand-in": {
			queueWith(object{"tiers": object{"basic": object{"profiles": object{"p.one": "p.two"}}}}), "granted",
		},
		"ungranted fallback": {queueWith(object{"fallbacks": object{"p.one": []string{"p.two"}}}), "granted"},
		"narrower fallback": {
			queueWith(object{"profiles": granted, "fallbacks": object{"p.one": []string{"p.narrow"}}}), "privacy",
		},
		"fallback elsewhere": {
			queueWith(object{"profiles": granted, "fallbacks": object{"p.one": []string{"p.other"}}}), "input_root",
		},
		"route to nowhere":   {queueWith(object{"runner": object{"profile": "p.gone"}}), "configured profiles"},
		"unnamed route":      {queueWith(object{"runner": object{"profile": ""}}), "runner.profile"},
		"shadow rate":        {shadow(0, "p.one", "ollama:model-a"), "shadow.rate"},
		"shadow no targets":  {shadow(1, "p.one"), "shadow.targets"},
		"shadow bare target": {shadow(1, "p.one", "ollama"), "shadow.targets"},
		"shadow paid router": {shadow(1, "p.one", "openrouter:openrouter/auto"), "shadow.targets"},
		"shadow no judge":    {shadow(1, "", "ollama:model-a"), "shadow.judge"},
		"shadow lost judge":  {shadow(1, "p.gone", "ollama:model-a"), "must be configured"},
		"shadow lost model":  {shadow(1, "p.one", "ollama:model-z"), "must be configured"},
		"shadow lost runner": {shadow(1, "p.one", "runner:p.gone"), "must be configured"},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := Load(writeConfig(t, test.data))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Errorf("got %v, want %q", err, test.want)
			}
		})
	}
}

func TestALoopbackListenHostIsAccepted(t *testing.T) {
	t.Parallel()

	for _, address := range []string{"localhost:8765", "[::1]:8765", "127.0.0.2:8765"} {
		config, err := Load(writeConfig(t, with("listen", address)))
		if err != nil || config.ListenPort != 8765 {
			t.Errorf("%s: %v", address, err)
		}
	}
}

func TestAMissingOrBrokenFileIsRefused(t *testing.T) {
	t.Parallel()

	_, err := Load(filepath.Join(t.TempDir(), "absent.json"))
	if err == nil {
		t.Error("an absent file loaded")
	}

	path := filepath.Join(t.TempDir(), "config.json")
	if os.WriteFile(path, []byte("{not json"), 0o600) != nil {
		t.Fatal("cannot write")
	}

	_, err = Load(path)
	if err == nil {
		t.Error("a broken file loaded")
	}
}
