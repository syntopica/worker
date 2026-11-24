package cli

// DrainsOnSignal reports whether the CLI command should trap termination
// so a long-running loop can release its resources and active work.
func DrainsOnSignal(name string) bool {
	switch name {
	case "serve", nodeCommand, tasksCommand, remoteCommand:
		return true
	default:
		return false
	}
}
