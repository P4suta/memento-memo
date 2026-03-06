package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/internal/repository"
	"github.com/sakashita/memento-memo/internal/service"
)

type SessionDetector struct {
	queries *repository.Queries
	pool    *pgxpool.Pool
}

func NewSessionDetector(pool *pgxpool.Pool) *SessionDetector {
	return &SessionDetector{
		queries: repository.New(pool),
		pool:    pool,
	}
}

// Run periodically purges old soft-deleted memos.
func (d *SessionDetector) Run(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run once immediately
	d.purge(ctx)

	for {
		select {
		case <-ticker.C:
			d.purge(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (d *SessionDetector) purge(ctx context.Context) {
	if err := d.queries.PurgeOldDeletedMemos(ctx); err != nil {
		slog.Error("failed to purge old deleted memos", slog.String("error", err.Error()))
	}
}

// GetFinalizedSessionDates returns dates that have finalized sessions.
func (d *SessionDetector) GetFinalizedSessionDates(ctx context.Context) ([]time.Time, error) {
	threshold := time.Now().Add(-service.SessionGapThreshold)
	sessions, err := d.queries.GetFinalizedSessions(ctx, service.ToPgTimestamptz(threshold))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var dates []time.Time
	for _, s := range sessions {
		dateStr := s.DateLabel.Time.Format("2006-01-02")
		if _, ok := seen[dateStr]; !ok {
			seen[dateStr] = struct{}{}
			dates = append(dates, s.DateLabel.Time)
		}
	}
	return dates, nil
}
