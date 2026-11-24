package jobs

import (
	"context"
	"strings"

	"github.com/syntopica/worker/internal/store"
)

func adminJobRows(ctx context.Context, s *store.Store, values params) ([]map[string]any, error) {
	rows, err := queryMapRows(ctx, s, adminJobQuery(values), values)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		row["sampling"] = sampling(row[keyProducer])
	}
	return rows, nil
}

func adminJobQuery(values params) string {
	var query strings.Builder
	query.WriteString(script("admin_job_select"))
	query.WriteString(script("admin_job_where"))
	for _, filter := range []struct {
		key, fragment string
	}{
		{keyID, "admin_job_id"},
		{keyQueue, "admin_job_queue"},
		{keyProducer, "admin_job_producer"},
		{"states", "admin_job_states"},
		{"before_id", "admin_job_before"},
	} {
		if values[filter.key] != nil {
			query.WriteString(script(filter.fragment))
		}
	}
	query.WriteString(script("admin_job_page"))
	return query.String()
}
