package jobs

import (
	"context"
	"errors"

	"github.com/syntopica/worker/internal/store"
)

const keyAgeSeconds = "age_s"

var errInvalidNodeReport = errors.New("jobs: invalid stored node report")

func statusNodes(ctx context.Context, s *store.Store, now float64) (map[string]map[string]any, error) {
	rows, err := queryMaps(ctx, s, "status_nodes", params{keyNow: now})
	if err != nil {
		return nil, err
	}
	nodes := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		value, decodeErr := decodeValue([]byte(text(row["report"])))
		report := object(value)
		if decodeErr != nil || report == nil {
			return nil, errInvalidNodeReport
		}
		report[keyAgeSeconds] = row[keyAgeSeconds]
		nodes[text(row["name"])] = report
	}
	releases, err := queryMaps(ctx, s, "status_releases", params{keyNow: now, keySince: now - daySeconds})
	if err != nil {
		return nil, err
	}
	for _, row := range releases {
		if report := nodes[text(row[keyNode])]; report != nil {
			report["last_release"] = map[string]any{keyCode: row[keyError], keyAgeSeconds: row[keyAgeSeconds]}
		}
	}
	return nodes, nil
}
