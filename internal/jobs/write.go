package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// write runs work in one write transaction and hands back what it produced.
func write[T any](ctx context.Context, s *store.Store, work func(tx store.Conn) (T, error)) (T, error) {
	var out T
	err := s.Write(ctx, func(tx store.Conn) error {
		value, inner := work(tx)
		out = value

		return inner
	})

	return out, err
}
