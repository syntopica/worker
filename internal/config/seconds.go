package config

import "time"

// seconds turns a configured number of seconds into a duration.
func seconds(value float64) time.Duration {
	return time.Duration(value * float64(time.Second))
}
