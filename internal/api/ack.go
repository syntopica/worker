package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/jobs"
)

func ack(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	b, err := decodeBody(r.raw, true)
	if err != nil {
		return 0, nil, err
	}
	rating, _ := b["rating"].(string)
	if b["rating"] != nil && !jobs.KnownRating(rating) {
		return 0, nil, fault.New(http.StatusBadRequest, "unknown_rating")
	}
	id, _ := b["result_id"].(string)
	err = jobs.AckResult(ctx, r.store, r.cfg, r.principal.Name, r.id, contract.Ack{
		ResultID: id, Rating: rating, Decline: truthy(b["decline"]),
	}, r.now)
	return http.StatusOK, map[string]bool{"acked": true}, err
}
