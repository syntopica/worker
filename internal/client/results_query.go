package client

import "time"

// ResultsQuery selects the unacknowledged results of one queue.
type ResultsQuery struct {
	// Queue is the queue to read.
	Queue string
	// After is the last Seq already seen; zero reads from the start.
	After int64
	// Limit bounds the page; zero means 50.
	Limit int
	// Wait long-polls for up to this long when nothing is waiting.
	Wait time.Duration
}
