package remote

import "context"

// Provider is the remote endpoint the loop spends its requests on.
type Provider interface {
	// Headroom is the free requests left today; false when unknown.
	Headroom(ctx context.Context) (left int, known bool)
	// Post sends one chat completion: an answer, or nil and an allowlisted
	// code.
	Post(ctx context.Context, body map[string]any) (answer *Answer, code string)
}
