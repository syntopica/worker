package config

import (
	"errors"
	"slices"
	"strings"
)

const defaultShadowPending = 20

var (
	errShadowRate   = errors.New("shadow.rate must be in (0, 1]")
	errShadowTarget = errors.New("shadow.targets must list kind:name executors, openrouter ones :free")
	errShadowJudge  = errors.New("shadow.judge must name a task profile")
)

// parseShadow builds a queue's configured shadow policy.
func parseShadow(raw *rawShadow) (*ShadowPolicy, error) {
	if raw.Rate <= 0 || raw.Rate > 1 {
		return nil, errShadowRate
	}

	if len(raw.Targets) == 0 || slices.ContainsFunc(raw.Targets, badTarget) {
		return nil, errShadowTarget
	}

	if raw.Judge == "" {
		return nil, errShadowJudge
	}

	return &ShadowPolicy{
		Rate:       raw.Rate,
		Targets:    raw.Targets,
		Judge:      raw.Judge,
		MaxPending: orDefault(raw.MaxPending, defaultShadowPending),
	}, nil
}

// badTarget reports a target that is not kind:name, or an OpenRouter one
// that is not free.
func badTarget(target string) bool {
	kind, name, _ := strings.Cut(target, ":")

	return !slices.Contains([]string{kindOllama, kindOpenRouter, kindRunner}, kind) || name == "" ||
		kind == kindOpenRouter && !freeModel(name)
}
