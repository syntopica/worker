package jobs

// jobRow is the part of a jobs row the lifecycle reads. A NULL text column
// reads as empty.
type jobRow struct {
	ID            string
	Producer      string
	Queue         string
	Kind          string
	Key           string
	Priority      int
	Privacy       string
	Model         string
	State         string
	Attempts      int
	MaxAttempts   int
	Preemptions   int
	Parked        bool
	ParkedRuns    int
	ParkedMinIdle float64
	Generation    int64
	LeaseNode     string
	LeaseAttempt  string
	Error         string
	Tier          string
	ShadowOf      string
	Pin           string
	Exclude       string
}
