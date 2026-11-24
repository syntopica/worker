package config

type rawConfig struct {
	Listen                *string               `json:"listen"`
	Models                map[string]rawModel   `json:"models"`
	Queues                map[string]rawQueue   `json:"queues"`
	Nodes                 map[string]rawNode    `json:"nodes"`
	Producers             map[string][]string   `json:"producers"`
	Privacy               map[string]rawClass   `json:"privacy"`
	MaxPayloadBytes       *int                  `json:"max_payload_bytes"`
	MaxSplit              *int                  `json:"max_split"`
	MaxParkedRuns         *int                  `json:"max_parked_runs"`
	SplitAfterPreemptions *int                  `json:"split_after_preemptions"`
	Profiles              map[string]rawProfile `json:"profiles"`
	Runners               map[string]rawRunner  `json:"runners"`
}
