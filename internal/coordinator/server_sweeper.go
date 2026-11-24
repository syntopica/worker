package coordinator

import (
	"context"
	"fmt"
	"time"

	"github.com/syntopica/worker/internal/jobs"
)

func (s *Server) runSweeper(ctx context.Context) {
	const sweepInterval = time.Minute
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()

	for ctx.Err() == nil {
		if err := s.sweep(ctx); err != nil && ctx.Err() == nil {
			_, _ = fmt.Fprintln(s.log, "worker: sweeper pass failed")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Server) sweep(ctx context.Context) error {
	const sweepTimeout = 45 * time.Second
	ctx, cancel := context.WithTimeout(ctx, sweepTimeout)
	defer cancel()

	cfg, now := s.source.Current(), epochSeconds()
	if _, err := jobs.ExpireLeases(ctx, s.store, now); err != nil {
		return err
	}
	if _, err := jobs.AdvanceShadowGroups(ctx, s.store, cfg, now); err != nil {
		return err
	}
	if _, err := jobs.SweepRetention(ctx, s.store, cfg, now); err != nil {
		return err
	}
	return s.store.Protect(ctx)
}
