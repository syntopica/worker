package link

// HostReport is the node's metadata-only host state for "worker status".
type HostReport struct {
	// Reason is why the node takes no work, or nil.
	Reason *string `json:"reason"`
	// IdleS is the idle time in seconds, nil when unreadable.
	IdleS *float64 `json:"idle_s"`
	// OnAC is the power source, nil when unreadable.
	OnAC *bool `json:"on_ac"`
	// Pressure is the memory pressure level.
	Pressure string `json:"pressure"`
	// Resident lists the models in memory.
	Resident []string `json:"resident"`
	// Unexpected lists the resident models the configuration does not manage.
	Unexpected []string `json:"unexpected"`
	// Loads1h counts the model loads of the last hour.
	Loads1h int `json:"loads_1h"`
	// Unloads1h counts the model unloads of the last hour.
	Unloads1h int `json:"unloads_1h"`
}
