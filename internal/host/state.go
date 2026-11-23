package host

import "time"

// State is one sample of the signals that decide whether a machine may work.
// A nil field means the read failed.
type State struct {
	// Idle is the time since the last keyboard or pointer input.
	Idle *time.Duration
	// OnAC says whether the machine is on AC power.
	OnAC *bool
	// Pressure is the memory pressure level.
	Pressure Pressure
	// Load1m is the one-minute load average, read only on a Linux server.
	Load1m *float64
}
