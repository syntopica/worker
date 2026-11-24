package jobs

type queueStatus struct {
	States        map[string]any `json:"states"`
	OldestQueued  *float64       `json:"oldest_queued_s"`
	Done          any            `json:"done_1h"`
	Wasted        any            `json:"wasted_1h_s"`
	SamplingFail  any            `json:"sampling_failed"`
	SamplingQueue any            `json:"sampling_queued"`
}
