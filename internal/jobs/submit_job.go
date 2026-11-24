package jobs

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/store"
)

// SubmitJob admits one job (spec 6): it validates, deduplicates, bounds and
// stores it, and reports the job's id and whether this call created it. Every
// refusal is a *fault.Error.
func SubmitJob(
	ctx context.Context, s *store.Store, cfg *config.Config, producer string, raw []byte, now float64,
) (id string, created bool, err error) {
	if len(raw) > cfg.MaxPayloadBytes {
		return "", false, fault.New(http.StatusRequestEntityTooLarge, "payload_too_large")
	}
	decoded, err := decodeValue(raw)
	body, isObject := decoded.(map[string]any)
	if err != nil || !isObject {
		return "", false, bad("bad_json")
	}
	sub, err := parseSubmit(body)
	if err != nil {
		return "", false, err
	}
	if grantErr := checkQueueGrant(cfg, producer, sub.Queue); grantErr != nil {
		return "", false, grantErr
	}
	model, err := resolveJobModel(cfg, sub)
	if err != nil {
		return "", false, err
	}
	digest := payloadHash(body)
	result, err := write(ctx, s, func(tx store.Conn) (contract.Submitted, error) {
		return submitOnce(ctx, tx, cfg, producer, sub, digest, model, now)
	})
	if err != nil {
		return "", false, err
	}
	return result.ID, result.Created, nil
}

func submitOnce(
	ctx context.Context, tx store.Conn, cfg *config.Config, producer string, sub *submission, digest, model string,
	now float64,
) (contract.Submitted, error) {
	var existing, stored string
	found, err := queryRow(ctx, tx, "existing_job",
		params{keyProducer: producer, keyQueue: sub.Queue, keyKey: sub.Key}, &existing, &stored)
	if err != nil {
		return contract.Submitted{}, err
	}
	if found {
		if stored != digest {
			return contract.Submitted{}, fault.New(http.StatusConflict, "idempotency_conflict")
		}
		return contract.Submitted{ID: existing}, nil
	}
	id, err := admit(ctx, tx, cfg, producer, sub, digest, model, now)
	return contract.Submitted{ID: id, Created: true}, err
}

// admit bounds and stores a job that is not a repeat of an earlier one.
func admit(
	ctx context.Context, tx store.Conn, cfg *config.Config, producer string, sub *submission, digest, model string,
	now float64,
) (string, error) {
	var err error
	if sub.ParentID != nil {
		err = admitSplitChild(ctx, tx, cfg, producer, sub, now)
	} else {
		err = checkOutstanding(ctx, tx, cfg, producer, sub.Queue)
	}
	if err != nil {
		return "", err
	}
	id := newID()
	if err := insertJob(ctx, tx, id, producer, sub, digest, model, "", now); err != nil {
		return "", err
	}

	return id, failEmptyLadder(ctx, tx, cfg, id, now)
}
