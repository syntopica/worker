package host

import "context"

// Sampler reads one State; a failed read degrades its field, never the call.
type Sampler func(ctx context.Context) State
