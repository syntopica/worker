package config

type rawQueue struct {
	RunWhen         *string             `json:"run_when"`
	Weight          *int                `json:"weight"`
	MaxOutstanding  *int                `json:"max_outstanding"`
	RetentionDays   *int                `json:"retention_days"`
	UnackedTTLHours *int                `json:"unacked_ttl_hours"`
	ParkedMinIdleS  *float64            `json:"parked_min_idle_s"`
	MaxModelAgeS    *float64            `json:"max_model_age_s"`
	Profiles        []string            `json:"profiles"`
	Fallbacks       map[string][]string `json:"fallbacks"`
	OpenRouter      *rawOpenRouter      `json:"openrouter"`
	Tiers           map[string]rawTier  `json:"tiers"`
	Runner          *rawRunnerRoute     `json:"runner"`
	LocalAfterS     float64             `json:"local_after_s"`
	Shadow          *rawShadow          `json:"shadow"`
}
