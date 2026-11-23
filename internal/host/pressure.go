package host

// Pressure is a named memory pressure level.
type Pressure string

const (
	// PressureNormal lets work run.
	PressureNormal Pressure = "normal"
	// PressureWarn releases work once it is sustained.
	PressureWarn Pressure = "warn"
	// PressureCritical always counts as pressure.
	PressureCritical Pressure = "critical"
	// PressureUnknown is a failed read, which blocks work.
	PressureUnknown Pressure = "unknown"
)

// Pressed reports whether the level is warn or critical.
func (p Pressure) Pressed() bool {
	return p == PressureWarn || p == PressureCritical
}
