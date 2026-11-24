package auth

// Principal is who is calling: Kind is producer, node or admin.
type Principal struct {
	Kind string
	Name string
}
