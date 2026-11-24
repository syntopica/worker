package jobs

import (
	"net/http"
	"slices"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/fault"
)

// checkQueueGrant refuses with queue_not_granted unless the producer may use
// the queue and the queue is configured.
func checkQueueGrant(cfg *config.Config, producer, queue string) error {
	if _, configured := cfg.Queues[queue]; !configured || !slices.Contains(cfg.Producers[producer], queue) {
		return fault.New(http.StatusForbidden, "queue_not_granted")
	}

	return nil
}
