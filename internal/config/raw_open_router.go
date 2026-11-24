package config

type rawOpenRouter struct {
	Models      map[string]string `json:"models"`
	AfterS      float64           `json:"after_s"`
	ZDR         *bool             `json:"zdr"`
	Fallbacks   []string          `json:"fallbacks"`
	DailyCap    *int              `json:"daily_cap"`
	DailyKeyCap *int              `json:"daily_key_cap"`
}
