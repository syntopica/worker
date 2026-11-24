package cli

import "slices"

func terminal(state string) bool {
	return slices.Contains([]string{"succeeded", "failed", "expired", "cancelled", "superseded", "unacked_expired", "reconciled"}, state)
}
