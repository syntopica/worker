package node

import "time"

// residencyEvent is one model load or unload the node saw.
type residencyEvent struct {
	at   time.Time
	load bool
}
