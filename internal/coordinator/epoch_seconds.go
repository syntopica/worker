package coordinator

import "time"

func epochSeconds() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Second)
}
