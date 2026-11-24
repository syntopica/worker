package config

type rawProfile struct {
	Runner    string            `json:"runner"`
	Model     string            `json:"model"`
	Reasoning string            `json:"reasoning"`
	Privacy   []string          `json:"privacy"`
	TimeoutS  *float64          `json:"timeout_s"`
	Nodes     []string          `json:"nodes"`
	InputRoot *string           `json:"input_root"`
	Command   string            `json:"command"`
	EnvUnset  []string          `json:"env_unset"`
	OnDemand  bool              `json:"on_demand"`
	Env       map[string]string `json:"env"`
}
