package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// queryMaps reads each row as column name to value, for the reports that are
// passed on as JSON unchanged.
func queryMaps(ctx context.Context, c store.Conn, name string, p params) ([]map[string]any, error) {
	return queryMapRows(ctx, c, script(name), p)
}
