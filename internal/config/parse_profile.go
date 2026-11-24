package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	defaultTimeoutS = 900
	maxLane         = "max-lane"
)

var (
	errRunner     = errors.New("runner must be one of codex, agy, cursor, max-lane, claude")
	errNoInputs   = errors.New("agy, claude and max-lane profiles take no input_root")
	errNamedModel = errors.New("claude and max-lane profiles name a model")
	errOnDemand   = errors.New("max-lane profiles are on_demand")
	errEnvName    = errors.New("env must map variable names to strings")
)

// parseProfile builds one task profile, applying the defaults.
//
// InputRoot resolves against the directory of config.json, the rule every
// Syntopica path follows. An agy profile takes no inputs: reading files
// needs a later and explicit decision. A max-lane profile has no tools
// either, pins its model, and is on demand: it spends a subscription the
// owner releases explicitly. A claude profile names its model and takes no
// inputs either, but runs on ordinary routes.
func parseProfile(name string, raw *rawProfile, configDir string) (TaskProfile, error) {
	profile := TaskProfile{
		Name:       name,
		Runner:     raw.Runner,
		Model:      raw.Model,
		Reasoning:  raw.Reasoning,
		Privacy:    raw.Privacy,
		Timeout:    seconds(orDefault(raw.TimeoutS, defaultTimeoutS)),
		Nodes:      raw.Nodes,
		Command:    raw.Command,
		DeniedRoot: resolve(filepath.Join(configDir, "state")),
		EnvUnset:   raw.EnvUnset,
		OnDemand:   raw.OnDemand,
	}

	if raw.Privacy == nil {
		profile.Privacy = []string{"public", "internal"}
	}

	err := checkProfile(raw)
	if _, unnamed := raw.Env[""]; unnamed && err == nil {
		err = errEnvName
	}

	if err != nil {
		return profile, fmt.Errorf("profile %s: %w", name, err)
	}

	if raw.InputRoot != nil {
		profile.InputRoot = resolve(filepath.Join(configDir, *raw.InputRoot))
		if filepath.IsAbs(*raw.InputRoot) {
			profile.InputRoot = resolve(*raw.InputRoot)
		}
	}

	profile.Env = profileEnv(raw.Env)

	return profile, nil
}

func checkProfile(raw *rawProfile) error {
	pinned := raw.Runner == maxLane || raw.Runner == "claude"

	switch {
	case !slices.Contains([]string{"codex", "agy", "cursor", maxLane, "claude"}, raw.Runner):
		return errRunner
	case raw.InputRoot != nil && (pinned || raw.Runner == "agy"):
		return errNoInputs
	case pinned && raw.Model == "":
		return errNamedModel
	case raw.Runner == maxLane && !raw.OnDemand:
		return errOnDemand
	}

	return nil
}

// profileEnv is the sorted pairs a runner is started with; a leading "~/"
// expands to the home directory.
func profileEnv(raw map[string]string) [][2]string {
	pairs := make([][2]string, 0, len(raw))

	for _, key := range slices.Sorted(maps.Keys(raw)) {
		value := raw[key]
		if rest, home := strings.CutPrefix(value, "~/"); home {
			directory, _ := os.UserHomeDir()
			value = filepath.Join(directory, rest)
		}

		pairs = append(pairs, [2]string{key, value})
	}

	return pairs
}
