package cli

// userError is a diagnostic whose message is safe to print verbatim.
type userError string

func (e userError) Error() string { return string(e) }
