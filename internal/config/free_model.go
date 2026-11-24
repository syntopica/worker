package config

import "strings"

// freeModel reports whether an OpenRouter model may be called without a
// budget: a ":free" endpoint, or the free router that picks one.
func freeModel(model string) bool {
	return strings.HasSuffix(model, ":free") || model == FreeRouterModel
}
