package jobs

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/store"
)

// RetryJob copies a finished job and acknowledges the original atomically.
func RetryJob(
	ctx context.Context, s *store.Store, cfg *config.Config, id, principal string, now float64,
) (contract.Submitted, error) {
	result, err := write(ctx, s, func(tx store.Conn) (contract.Submitted, error) {
		job, err := ownedJob(ctx, tx, nil, id)
		if err != nil {
			return contract.Submitted{}, err
		}
		if job.State != stateFailed && job.State != stateExpired && job.State != stateCancelled {
			return contract.Submitted{}, fault.New(http.StatusConflict, "not_retryable")
		}
		return prepareRetry(ctx, tx, cfg, job, principal, now)
	})
	if err != nil {
		return contract.Submitted{}, err
	}
	return result, nil
}

func prepareRetry(
	ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, principal string, now float64,
) (contract.Submitted, error) {
	input, err := loadJobInput(ctx, tx, job.ID)
	if err != nil {
		return contract.Submitted{}, err
	}
	if input == nil {
		return contract.Submitted{}, fault.New(http.StatusGone, "content_gone")
	}
	body, err := retryBody(job, input)
	if err != nil {
		return contract.Submitted{}, err
	}
	sub, err := parseSubmit(body)
	if err != nil {
		return contract.Submitted{}, err
	}
	if grantErr := checkQueueGrant(cfg, job.Producer, sub.Queue); grantErr != nil {
		return contract.Submitted{}, grantErr
	}
	model, err := resolveJobModel(cfg, sub)
	if err != nil {
		return contract.Submitted{}, err
	}
	existing, found, err := retryExisting(ctx, tx, job, sub.Key)
	if err != nil || found {
		return contract.Submitted{ID: existing}, err
	}
	id, err := retryCopy(ctx, tx, cfg, job, sub, payloadHash(body), model, now)
	if err != nil {
		return contract.Submitted{}, err
	}
	err = recordAudit(ctx, tx, job.ID, job.Privacy, principal, "retry", now)
	return contract.Submitted{ID: id, Created: true}, err
}

func retryExisting(ctx context.Context, tx store.Conn, job *jobRow, key string) (id string, found bool, err error) {
	var digest string
	found, err = queryRow(ctx, tx, "existing_job", params{keyProducer: job.Producer, keyQueue: job.Queue, keyKey: key},
		&id, &digest)
	return id, found, err
}
