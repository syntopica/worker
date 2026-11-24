package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	errGranted      = errors.New("fallbacks and stand-ins must name granted profiles")
	errWiderPrivacy = errors.New("a stand-in must accept its primary's privacy")
	errOtherRoot    = errors.New("a stand-in must share its primary's input_root")
	errTierModel    = errors.New("tiers name unknown models")
	errRouteProfile = errors.New("runner profiles must be configured profiles")
	errShadowKnown  = errors.New("shadow targets and judge must be configured")
)

// checkRoutes refuses a queue whose routes name something unconfigured or
// would widen what a task may reach.
func checkRoutes(config *Config) error {
	for name := range config.Queues {
		queue := config.Queues[name]

		err := checkQueue(config, &queue)
		if err != nil {
			return fmt.Errorf("queue %s: %w", name, err)
		}
	}

	return nil
}

func checkQueue(config *Config, queue *QueuePolicy) error {
	err := checkStandIns(config, queue)
	if err != nil {
		return err
	}

	for _, route := range queue.Tiers {
		for _, model := range route.Models {
			if _, known := config.Models[model]; !known {
				return errTierModel
			}
		}
	}

	if route := queue.Runner; route != nil {
		for _, name := range append([]string{route.Profile}, route.Fallbacks...) {
			if _, known := config.Profiles[name]; !known {
				return errRouteProfile
			}
		}
	}

	return checkShadow(config, queue.Shadow)
}

func checkStandIns(config *Config, queue *QueuePolicy) error {
	for primary, alternatives := range queue.Fallbacks {
		for _, other := range append([]string{primary}, alternatives...) {
			err := checkStandIn(config, queue, primary, other)
			if err != nil {
				return err
			}
		}
	}

	for _, route := range queue.Tiers {
		for primary, other := range route.Profiles {
			err := checkStandIn(config, queue, primary, other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// checkStandIn refuses a profile that would widen what a task admitted
// under another may reach.
//
// It must accept every privacy class the primary accepts (the job was
// admitted against the primary) and read from the same input root, so the
// same manifest means the same files.
func checkStandIn(config *Config, queue *QueuePolicy, primary, other string) error {
	base, known := config.Profiles[primary]
	stand, alsoKnown := config.Profiles[other]

	if !known || !alsoKnown || !slices.Contains(queue.Profiles, primary) || !slices.Contains(queue.Profiles, other) {
		return errGranted
	}

	for _, class := range base.Privacy {
		if !slices.Contains(stand.Privacy, class) {
			return errWiderPrivacy
		}
	}

	if base.InputRoot != "" && stand.InputRoot != base.InputRoot {
		return errOtherRoot
	}

	return nil
}

func checkShadow(config *Config, shadow *ShadowPolicy) error {
	if shadow == nil {
		return nil
	}

	if _, known := config.Profiles[shadow.Judge]; !known {
		return errShadowKnown
	}

	for _, target := range shadow.Targets {
		kind, name, _ := strings.Cut(target, ":")
		_, model := config.Models[name]
		_, profile := config.Profiles[name]

		if kind == kindOllama && !model || kind == kindRunner && !profile {
			return errShadowKnown
		}
	}

	return nil
}
