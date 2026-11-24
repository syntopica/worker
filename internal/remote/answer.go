package remote

import "github.com/syntopica/worker/internal/jsonanswer"

// Answer is a finished chat completion.
type Answer struct {
	// Model is the model that answered, empty when the endpoint named none.
	Model string
	// Content is the first choice's message content.
	Content string
	// TokensIn is the prompt token count.
	TokensIn int
	// TokensOut is the generated token count.
	TokensOut int
	// CostUSD is what the endpoint charged.
	CostUSD float64
}

// Output is the contract's output: the text, and the text as JSON when it is
// a JSON document, else null. A single surrounding code fence is tolerated:
// some free models add one even under a JSON response format.
func (a *Answer) Output() map[string]any {
	return map[string]any{"text": a.Content, "json": jsonanswer.Parse(a.Content)}
}

// Usage is the contract's usage of the call.
func (a *Answer) Usage() map[string]any {
	return map[string]any{"tokens_in": a.TokensIn, "tokens_out": a.TokensOut, "cost_usd": a.CostUSD}
}
