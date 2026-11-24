package api

import (
	"encoding/json"
	"net/http"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/policy"
)

func leaseRequest(r *request, b body) (*policy.LeaseRequest, error) {
	kind, profile, err := leaseRoute(r.cfg, b)
	if err != nil {
		return nil, err
	}
	free, err := b.number("free_gb", 0)
	if err != nil {
		return nil, err
	}
	idle, err := numberValue(b["current_idle_s"])
	if err != nil {
		return nil, err
	}
	active, found := b["user_active"]
	if !found {
		return nil, badRequest()
	}
	model, _ := b["resident_model"].(string)
	cutoff, err := leaseCutoff(b["created_before"], profile)
	if err != nil {
		return nil, err
	}
	override, err := pinOverride(b, profile)
	if err != nil {
		return nil, err
	}
	return &policy.LeaseRequest{
		Node: r.principal.Name, Kind: kind, Profile: profile, FreeGB: free, CurrentIdle: idle,
		UserActive: truthy(active), ResidentModel: model, CreatedBefore: cutoff, OverridePins: override,
	}, nil
}

func leaseRoute(cfg *config.Config, b body) (kind, profile string, err error) {
	const inference, task, openRouter = "inference", "task", "openrouter"
	kind = inference
	if value, found := b["kind"]; found {
		kind, _ = value.(string)
	}
	if kind != inference && kind != task && kind != openRouter {
		return "", "", fault.New(http.StatusBadRequest, "unsupported_kind")
	}
	if value := b["profile"]; value != nil {
		profile, _ = value.(string)
		known, found := cfg.Profiles[profile]
		if kind != task || !found || !known.OnDemand {
			return "", "", fault.New(http.StatusBadRequest, "unknown_profile")
		}
	}
	return kind, profile, nil
}

func leaseCutoff(value any, profile string) (*float64, error) {
	if value == nil {
		var absent *float64
		return absent, nil
	}
	number, ok := value.(json.Number)
	if !ok || profile == "" {
		return nil, badRequest()
	}
	seconds, err := numberValue(number)
	if err != nil || seconds < 0 {
		return nil, badRequest()
	}
	return &seconds, nil
}

func pinOverride(b body, profile string) (bool, error) {
	value, found := b["override_pins"]
	if !found {
		return false, nil
	}
	override, ok := value.(bool)
	if !ok || (override && profile == "") {
		return false, badRequest()
	}
	return override, nil
}
