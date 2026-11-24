package jobs

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/store"
)

// ReadAdminContent reads payloads only after checking the reveal class and recording any required audit.
func ReadAdminContent(ctx context.Context, s *store.Store, id, reveal, principal string, now float64) (map[string]any, error) {
	content, err := write(ctx, s, func(tx store.Conn) (map[string]any, error) {
		job, err := ownedJob(ctx, tx, nil, id)
		if err != nil {
			return nil, err
		}
		sensitive := Sensitive(job.Privacy)
		if sensitive && reveal != job.Privacy {
			return nil, fault.New(http.StatusForbidden, "reveal_required")
		}
		content, err := loadContent(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if sensitive {
			if err := recordAudit(ctx, tx, id, job.Privacy, principal, "reveal", now); err != nil {
				return nil, err
			}
		}
		return content, nil
	})
	if err != nil {
		return nil, err
	}
	return content, nil
}

func loadContent(ctx context.Context, tx store.Conn, id string) (map[string]any, error) {
	input, err := loadJobInput(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	output, err := loadJobOutput(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if input == nil && output == nil {
		return nil, fault.New(http.StatusGone, "content_gone")
	}
	return map[string]any{keyInput: input, "output": output}, nil
}
