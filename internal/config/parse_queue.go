package config

import (
	"errors"
	"fmt"
	"slices"
)

const (
	defaultOutstanding   = 200
	defaultRetentionDays = 7
	defaultUnackedHours  = 72
	defaultParkedIdleS   = 600
	defaultModelAgeS     = 3600
	hoursPerDay          = 24
)

var (
	errRunWhen     = errors.New("run_when must be idle or active_ok")
	errRetention   = errors.New("retention_days must be >= 1 and cover unacked_ttl_hours")
	errTier        = errors.New("tiers must be basic or strong")
	errRunnerRoute = errors.New("runner.profile and runner.fallbacks must name task profiles")
)

// parseQueue builds one queue policy, applying the spec defaults.
//
// RetentionDays must cover UnackedTTLHours: a shorter one would sweep a
// failed control result before its producer fetched it.
func parseQueue(name string, raw *rawQueue) (QueuePolicy, error) {
	queue := QueuePolicy{
		Name:            name,
		RunWhen:         orDefault(raw.RunWhen, "idle"),
		Weight:          orDefault(raw.Weight, 1),
		MaxOutstanding:  orDefault(raw.MaxOutstanding, defaultOutstanding),
		RetentionDays:   orDefault(raw.RetentionDays, defaultRetentionDays),
		UnackedTTLHours: orDefault(raw.UnackedTTLHours, defaultUnackedHours),
		ParkedMinIdle:   seconds(orDefault(raw.ParkedMinIdleS, defaultParkedIdleS)),
		MaxModelAge:     seconds(orDefault(raw.MaxModelAgeS, defaultModelAgeS)),
		Profiles:        raw.Profiles,
		Fallbacks:       raw.Fallbacks,
		LocalAfter:      seconds(raw.LocalAfterS),
	}

	err := fillRoutes(&queue, raw)
	if err != nil {
		return queue, fmt.Errorf("queue %s: %w", name, err)
	}

	return queue, nil
}

func fillRoutes(queue *QueuePolicy, raw *rawQueue) error {
	if err := checkQueueLimits(queue); err != nil {
		return err
	}

	var err error

	queue.Tiers, err = tierRoutes(raw.Tiers)
	if err != nil {
		return err
	}

	if raw.Runner != nil {
		queue.Runner, err = runnerRoute(raw.Runner)
		if err != nil {
			return err
		}
	}

	if raw.OpenRouter != nil {
		queue.OpenRouter, err = parseOpenRouter(raw.OpenRouter)
		if err != nil {
			return err
		}
	}

	if raw.Shadow != nil {
		queue.Shadow, err = parseShadow(raw.Shadow)
	}

	return err
}

func checkQueueLimits(queue *QueuePolicy) error {
	if queue.RunWhen != "idle" && queue.RunWhen != "active_ok" {
		return errRunWhen
	}
	if queue.RetentionDays < 1 || queue.RetentionDays*hoursPerDay < queue.UnackedTTLHours {
		return errRetention
	}
	return nil
}

func tierRoutes(raw map[string]rawTier) (map[string]TierRoute, error) {
	tiers := make(map[string]TierRoute, len(raw))

	for tier, entry := range raw {
		if tier != "basic" && tier != "strong" {
			return nil, errTier
		}

		tiers[tier] = TierRoute(entry)
	}

	return tiers, nil
}

func runnerRoute(raw *rawRunnerRoute) (*RunnerRoute, error) {
	if raw.Profile == "" || slices.Contains(raw.Fallbacks, "") {
		return nil, errRunnerRoute
	}

	return &RunnerRoute{Profile: raw.Profile, After: seconds(raw.AfterS), Fallbacks: raw.Fallbacks}, nil
}
