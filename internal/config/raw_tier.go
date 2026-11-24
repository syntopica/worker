package config

type rawTier struct {
	Models   []string          `json:"models"`
	Profiles map[string]string `json:"profiles"`
}
