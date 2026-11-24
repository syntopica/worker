package jobs

import (
	"context"
	"strconv"
	"strings"

	"github.com/syntopica/worker/internal/store"
)

const maxAdminJobs = 100

// ListAdminJobs returns job metadata in newest-first order with an opaque next cursor.
func ListAdminJobs(ctx context.Context, s *store.Store, wanted Filter) (map[string]any, error) {
	limit := min(max(wanted.Limit, 1), maxAdminJobs)
	values, err := adminFilters(wanted, limit)
	if err != nil {
		return nil, err
	}
	rows, err := adminJobRows(ctx, s, values)
	if err != nil {
		return nil, err
	}
	var next any
	if len(rows) == limit {
		last := rows[len(rows)-1]
		when, _ := last["created"].(float64)
		next = shortestFloat(when) + ":" + text(last[keyID])
	}
	return map[string]any{"jobs": rows, "next": next}, nil
}

func setBefore(values params, cursor string) error {
	stamp, id, _ := strings.Cut(cursor, ":")
	created, err := strconv.ParseFloat(stamp, 64)
	if err != nil || id == "" {
		return bad("bad_cursor")
	}
	values["before_time"], values["before_id"] = created, id
	return nil
}

func adminFilters(wanted Filter, limit int) (params, error) {
	values := adminParams(limit)
	if wanted.Queue != nil {
		values[keyQueue] = *wanted.Queue
	}
	if wanted.Producer != nil {
		values[keyProducer] = *wanted.Producer
	}
	if wanted.States != nil {
		states, err := encodeValue(wanted.States)
		if err != nil {
			return nil, err
		}
		values["states"] = states
	}
	if wanted.Before != nil {
		if err := setBefore(values, *wanted.Before); err != nil {
			return nil, err
		}
	}
	return values, nil
}
