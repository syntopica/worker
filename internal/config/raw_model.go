package config

type rawModel struct {
	NumCtx    *int     `json:"num_ctx"`
	KeepAlive *string  `json:"keep_alive"`
	ColdGB    *float64 `json:"cold_gb"`
	WarmGB    *float64 `json:"warm_gb"`
}

func (m rawModel) complete() bool {
	return m.NumCtx != nil && m.KeepAlive != nil && m.ColdGB != nil && m.WarmGB != nil
}
