package node

// reason is an allowlisted word: why the node takes no work, why an attempt
// was released, or why it failed.
type reason string

const (
	hostStateUnreadable reason = "host_state_unreadable"
	onBattery           reason = "on_battery"
	memoryPressure      reason = "memory_pressure"
	userActive          reason = "user_active"
	pressureRecovering  reason = "pressure_recovering"
	drainBlocked        reason = "drain_failed"
	backendDown         reason = "backend_down"
	pressureBackoff     reason = "pressure_backoff"
	transportBackoff    reason = "transport_backoff"
	transportError      reason = "transport_error"
	nodeShutdown        reason = "node_shutdown"
	nodeError           reason = "node_error"
	privacyRefused      reason = "privacy_refused"
	unknownModel        reason = "unknown_model"
)

// pointer is the reason as a report field: nil when there is none.
func (r reason) pointer() *string {
	if r == "" {
		return nil
	}

	return new(string(r))
}
