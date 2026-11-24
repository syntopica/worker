package jobs

import "github.com/syntopica/worker/internal/config"

// completionExecutor takes the node from the lease and keeps executor names
// from config, plus bounded model identifiers returned by the remote router.
func completionExecutor(cfg *config.Config, job *jobRow, reported map[string]any) map[string]any {
	executor := map[string]any{keyNode: job.LeaseNode}
	provider, _ := reported[keyProvider].(string)
	model, hasModel := reported[keyModel].(string)
	known, matches := completionModel(cfg, job, provider, model)
	if known {
		executor[keyProvider] = provider
	}
	if matches && hasModel {
		executor[keyModel] = model
	}
	return executor
}
