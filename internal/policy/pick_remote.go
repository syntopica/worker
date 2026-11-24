package policy

import (
	"strings"

	"github.com/syntopica/worker/internal/config"
)

// PickRemote chooses the next inference job to escalate to OpenRouter (spec
// 8, rung 3): one whose queue routes its model there and that waited long
// enough, by priority, weighted share and age. The pick carries the remote
// model id in Model; the job keeps its own.
//
// The route's After applies only while runnerFirst says a runner would take
// the job; otherwise it escalates at once. A job pinned to an OpenRouter
// model goes there at once; other pins never.
func PickRemote(
	candidates []Candidate, cfg *config.Config, recent map[string]int, now float64, runnerFirst func(*Candidate) bool,
) (Candidate, bool) {
	var eligible []Candidate
	for index := range candidates {
		remote, routed := remoteModel(&candidates[index], cfg, now, runnerFirst)
		if routed {
			eligible = append(eligible, candidates[index].under(remote))
		}
	}
	chosen := fairest(eligible, cfg, recent)
	if chosen < 0 {
		return Candidate{}, false
	}
	return eligible[chosen], true
}

// remoteModel is the OpenRouter model one candidate would run on now.
func remoteModel(c *Candidate, cfg *config.Config, now float64, runnerFirst func(*Candidate) bool) (string, bool) {
	if c.Pin != "" {
		kind, pinned, _ := strings.Cut(c.Pin, ":")
		return pinned, kind == kindOpenRouter
	}
	route := cfg.Queues[c.Queue].OpenRouter
	if route == nil || (now-c.Created < route.After.Seconds() && runnerFirst != nil && runnerFirst(c)) {
		return "", false
	}
	remote, mapped := route.Models[c.Model]
	return remote, mapped
}
