package policy

import (
	"slices"
	"strings"

	"github.com/syntopica/worker/internal/config"
)

// OnDemandCandidates is the queued work an explicit on-demand profile may
// take for the asking node: inference, prompt-only tasks and judges, each
// carrying the profile in Model.
//
// The profile must be on demand, rested and allowed on the node; each job's
// class must be allowed by the profile and for runner executors. Comparison
// pins stay with their submitted executors unless overridePins asks.
func OnDemandCandidates(
	candidates []Candidate, cfg *config.Config, cooling map[string]bool, node Node, name string, overridePins bool,
) []Candidate {
	profile, known := cfg.Profiles[name]
	if !known || !profile.OnDemand || ProfileResting(&profile, cooling) || !onNode(&profile, node.Name) {
		return nil
	}
	var kept []Candidate
	for index := range candidates {
		c := &candidates[index]
		overridable := onDemandInference(c, overridePins) || onDemandTask(c, cfg, node.Name)
		if overridable && onDemandGranted(c, cfg, &profile, node.Trust) {
			kept = append(kept, c.under(name))
		}
	}
	return kept
}

// onDemandGranted reports whether the profile and the runner executors take
// the job's class and the job does not exclude the profile's runner.
func onDemandGranted(c *Candidate, cfg *config.Config, profile *config.TaskProfile, trust string) bool {
	granted := slices.Contains(profile.Privacy, c.Privacy) && !slices.Contains(c.ExcludeProviders, profile.Runner)
	return granted && PrivacyAllows(cfg, c.Privacy, kindRunner, trust)
}

// onDemandInference keeps comparison pins protected unless their
// substitution was asked for; original-answer markers never become runnable.
func onDemandInference(c *Candidate, overridePins bool) bool {
	if c.Kind != kindInference || c.Pin == "" {
		return c.Kind == kindInference
	}
	kind, _, found := strings.Cut(c.Pin, ":")
	return overridePins && found && slices.Contains([]string{kindOllama, kindOpenRouter, kindRunner}, kind)
}

// onDemandTask lets only ordinary tasks and judges be overridden, within
// their source profile's grant.
func onDemandTask(c *Candidate, cfg *config.Config, node string) bool {
	source, known := cfg.Profiles[c.Model]
	if c.Kind != kindTask || (c.Pin != "" && c.Pin != pinJudge) || !known {
		return false
	}
	return !source.OnDemand && slices.Contains(source.Privacy, c.Privacy) && onNode(&source, node)
}
