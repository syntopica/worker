package auth

// entry is one principal in principals.json: its kind and the SHA-256 of
// its token, never the token.
type entry struct {
	Kind   string `json:"kind"`
	SHA256 string `json:"sha256"`
}
