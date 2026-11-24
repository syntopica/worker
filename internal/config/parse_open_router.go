package config

import (
	"errors"
	"slices"
)

var (
	errRouteModels    = errors.New("openrouter.models must map local to :free remote models")
	errRouteFallbacks = errors.New("openrouter.fallbacks must list :free endpoints")
	errRouteCap       = errors.New("openrouter daily caps must be positive integers")
)

// parseOpenRouter builds a queue's configured OpenRouter route.
//
// Only ":free" model ids and the free router are accepted: paid endpoints
// need a budget, which this route does not carry.
func parseOpenRouter(raw *rawOpenRouter) (*OpenRouterRoute, error) {
	if len(raw.Models) == 0 {
		return nil, errRouteModels
	}

	for _, remote := range raw.Models {
		if !freeModel(remote) {
			return nil, errRouteModels
		}
	}

	if slices.ContainsFunc(raw.Fallbacks, func(model string) bool { return !freeModel(model) }) {
		return nil, errRouteFallbacks
	}

	for _, limit := range []*int{raw.DailyCap, raw.DailyKeyCap} {
		if limit != nil && *limit < 1 {
			return nil, errRouteCap
		}
	}

	return &OpenRouterRoute{
		Models:      raw.Models,
		After:       seconds(raw.AfterS),
		ZDR:         orDefault(raw.ZDR, true),
		Fallbacks:   raw.Fallbacks,
		DailyCap:    raw.DailyCap,
		DailyKeyCap: raw.DailyKeyCap,
	}, nil
}
