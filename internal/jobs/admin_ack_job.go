package jobs

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/store"
)

// AdminAckJob acknowledges a job's open control results on its producer's behalf.
func AdminAckJob(ctx context.Context, s *store.Store, cfg *config.Config, id string, now float64) error {
	return s.Write(ctx, func(tx store.Conn) error {
		job, err := ownedJob(ctx, tx, nil, id)
		if err != nil {
			return err
		}
		return ackOnlyControls(ctx, tx, cfg, job, now)
	})
}

func ackOnlyControls(ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, now float64) error {
	type controlRow struct {
		id      string
		control *string
	}
	rows, err := queryAll(ctx, tx, "open_controls", params{keyID: job.ID}, func(row *sql.Rows) (controlRow, error) {
		var result controlRow
		err := scanInto(row, &result.id, &result.control)
		return result, err
	})
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return fault.New(http.StatusConflict, "not_ackable")
	}
	for _, row := range rows {
		if row.control == nil {
			return fault.New(http.StatusConflict, "not_ackable")
		}
		if err := applyAck(ctx, tx, cfg, job, contract.Ack{ResultID: row.id}, row.control, now); err != nil {
			return err
		}
	}
	return nil
}
