package contract

// Usage is what an attempt measured.
type Usage struct {
	// TokensIn counts prompt tokens.
	TokensIn int `json:"tokens_in"`
	// TokensOut counts generated tokens.
	TokensOut int `json:"tokens_out"`
}
