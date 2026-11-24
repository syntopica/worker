package api

import (
	"encoding/json"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/link"
)

func decodeHostReport(raw []byte, cfg *config.Config) (*link.HostReport, error) {
	fields, err := decodeBody(raw, false)
	if err != nil {
		return nil, err
	}
	for name, value := range fields {
		if !validHostReportField(name, value) {
			return nil, badRequest()
		}
	}
	report := link.HostReport{Pressure: string(host.PressureUnknown), Resident: []string{}, Unexpected: []string{}}
	if err := json.Unmarshal(raw, &report); err != nil || !validHostReport(&report) {
		return nil, badRequest()
	}
	for i, model := range report.Resident {
		report.Resident[i] = cfg.ModelMetadataName(model)
	}
	for i, model := range report.Unexpected {
		report.Unexpected[i] = cfg.ModelMetadataName(model)
	}

	return &report, nil
}

func validHostReport(report *link.HostReport) bool {
	return validHostReason(report.Reason) &&
		(report.IdleS == nil || *report.IdleS >= 0) &&
		report.Loads1h >= 0 && report.Unloads1h >= 0
}

func validHostReportField(name string, value any) bool {
	switch name {
	case "reason", "idle_s", "on_ac":
		return true
	case "pressure":
		pressure, ok := value.(string)
		return ok && validHostPressure(host.Pressure(pressure))
	case "loads_1h", "unloads_1h":
		return value != nil
	case "resident", "unexpected":
		return isModelList(value)
	default:
		return false
	}
}

func isModelList(value any) bool {
	models, ok := value.([]any)
	if !ok {
		return false
	}
	for _, model := range models {
		if _, ok := model.(string); !ok {
			return false
		}
	}
	return true
}

func validHostReason(reason *string) bool {
	if reason == nil {
		return true
	}
	switch *reason {
	case "host_state_unreadable", "on_battery", "memory_pressure", "user_active",
		"pressure_recovering", "drain_failed", "backend_down", "pressure_backoff", "transport_backoff":
		return true
	default:
		return false
	}
}

func validHostPressure(pressure host.Pressure) bool {
	switch pressure {
	case host.PressureNormal, host.PressureWarn, host.PressureCritical, host.PressureUnknown:
		return true
	default:
		return false
	}
}
