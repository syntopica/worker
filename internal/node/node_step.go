package node

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/link"
)

// step is one iteration of the node loop: sample, report, and maybe lease
// and run. It reports whether the loop should rest before the next one.
//
// A failed drain clears once /api/ps no longer lists its model, checked only
// while pressure is normal; no chat probe is ever sent for it. While it is
// still listed, its unload is retried on a slow timer.
func (n *Node) step(ctx context.Context) (rest bool, err error) {
	cfg := n.Config()

	policy, known := cfg.Nodes[n.Name]
	if !known {
		return true, errUnknownNode
	}

	m, backend := n.remembered(), n.Backend(policy.OllamaURL)
	state, now := n.Sample(ctx), n.Now()
	observeHostState(m, state, now)

	resident, answered := backend.Resident(ctx)
	if answered {
		maps.DeleteFunc(m.owned, func(model string, _ bool) bool { return !slices.Contains(resident, model) })

		if state.Pressure == host.PressureNormal && !slices.Contains(resident, m.failedModel) {
			m.failedModel = ""
		}

		retryFailedDrain(ctx, backend, m, resident, now)
	}

	idle := idleTime(state)
	ownerActive := policy.LocalIdleOnly && idle < policy.IdleThreshold
	relieve(ctx, backend, m, resident, state.OnAC, ownerActive)

	block := nodeBlock(m, state, now, !answered)
	n.Link.Report(ctx, n.hostReport(cfg, state, resident, answered, block, now))

	if block != "" || ownerActive {
		return true, nil
	}

	request := link.LeaseRequest{
		Node:         n.Name,
		UserActive:   idle < policy.IdleThreshold,
		FreeGB:       policy.MemoryBudgetGB,
		CurrentIdleS: idle.Seconds(),
	}

	if at := slices.IndexFunc(resident, func(model string) bool { return managed(cfg, model) }); at >= 0 {
		request.ResidentModel = &resident[at]
		request.FreeGB -= cfg.Models[resident[at]].ColdGB
	}

	lease := n.Link.Lease(ctx, request)
	if lease == nil {
		return true, nil
	}

	return n.runLeased(ctx, cfg, &policy, lease, slices.Contains(resident, lease.Model)), nil
}

// hostReport is the node's report of this iteration. Residency is tracked
// only when the backend answered.
func (n *Node) hostReport(
	cfg *config.Config, state host.State, resident []string, answered bool, block reason, now time.Time,
) *link.HostReport {
	report := &link.HostReport{
		Reason:     block.pointer(),
		OnAC:       state.OnAC,
		Pressure:   string(state.Pressure),
		Resident:   append([]string{}, resident...),
		Unexpected: []string{},
	}

	if state.Idle != nil {
		report.IdleS = new(state.Idle.Seconds())
	}

	for _, model := range resident {
		if !managed(cfg, model) {
			report.Unexpected = append(report.Unexpected, model)
		}
	}

	if answered {
		report.Loads1h, report.Unloads1h = trackResidency(n.remembered(), resident, now, n.Log)
	}

	return report
}
