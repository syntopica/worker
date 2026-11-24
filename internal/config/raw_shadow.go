package config

type rawShadow struct {
	Rate       float64  `json:"rate"`
	Targets    []string `json:"targets"`
	Judge      string   `json:"judge"`
	MaxPending *int     `json:"max_pending"`
}
