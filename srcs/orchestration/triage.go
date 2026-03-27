package orchestration

import (
	"context"
	"log/slog"
	"time"
)

// PruneStaleMissions removes missions that are older than the specified duration.
func (s *SIPDB) PruneStaleMissions(ctx context.Context, age time.Duration) (int64, error) {
	cutoff := time.Now().Add(-age)

	var affected int64
	err := withRetry(ctx, func() error {
		res, err := s.db.ExecContext(ctx, "DELETE FROM agent_missions WHERE status = 'COMPLETED' AND updated_at < ?", cutoff)
		if err != nil {
			return err
		}
		affected, err = res.RowsAffected()
		return err
	})

	if err == nil {
		slog.Info("Pruned stale missions", "count", affected, "cutoff", cutoff)
	}
	return affected, err
}
