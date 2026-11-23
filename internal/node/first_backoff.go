package node

import "time"

// firstBackoff is the first backoff after a pressure release.
const firstBackoff = 15 * time.Minute
