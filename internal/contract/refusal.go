package contract

// Refusal is the body of every non-2xx answer.
type Refusal struct {
	// Error is a code from the contract's fixed vocabulary.
	Error string `json:"error"`
}
