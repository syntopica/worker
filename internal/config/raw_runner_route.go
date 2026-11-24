package config

type rawRunnerRoute struct {
	Profile   string   `json:"profile"`
	AfterS    float64  `json:"after_s"`
	Fallbacks []string `json:"fallbacks"`
}
