package config

// ModelPin holds the reload-sensitive options one managed model is always
// called with.
type ModelPin struct {
	// Name is the model's name in the model server.
	Name string
	// NumCtx is the context window; it never comes from a job.
	NumCtx int
	// KeepAlive is how long the model stays resident after a request.
	KeepAlive string
	// ColdGB covers weights, KV for NumCtx x parallel, buffers and margin.
	ColdGB float64
	// WarmGB is what a job costs once the model is resident.
	WarmGB float64
}
