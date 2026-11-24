package config

import "errors"

var (
	errModelRequired   = errors.New("config: model requires num_ctx, keep_alive, cold_gb and warm_gb")
	errNodeRequired    = errors.New("config: node requires trust, memory_budget_gb, ollama_url and ollama_launchd_label")
	errPrivacyRequired = errors.New("config: privacy override requires executors and trust")
)

func checkRequiredFields(raw *rawConfig) error {
	for _, model := range raw.Models {
		if !model.complete() {
			return errModelRequired
		}
	}
	for _, node := range raw.Nodes {
		if !node.complete() {
			return errNodeRequired
		}
	}
	for _, class := range raw.Privacy {
		if class.Executors == nil || class.Trust == nil {
			return errPrivacyRequired
		}
	}
	return nil
}
