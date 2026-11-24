package config

type rawRunner struct {
	CooldownS     *float64            `json:"cooldown_s"`
	QuotaProvider string              `json:"quota_provider"`
	QuotaWindows  []string            `json:"quota_windows"`
	ModelWindows  map[string][]string `json:"model_windows"`
	MaxConcurrent *int                `json:"max_concurrent"`
}
