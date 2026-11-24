package jobs

// Filter contains validated filters for the administrative job list.
type Filter struct {
	Queue    *string
	States   []string
	Producer *string
	Before   *string
	Limit    int
}
