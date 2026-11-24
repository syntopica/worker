package cli

import "github.com/syntopica/worker/internal/contract"

func formatExecutor(executor *contract.Executor) string {
	if executor == nil {
		return unknownValue + " " + unknownValue
	}
	provider, model := executor.Provider, executor.Model
	if provider == "" {
		provider = unknownValue
	}
	if model == "" {
		model = unknownValue
	}
	return provider + " " + model
}
