package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/internal/service"
)

type ReportGenerator struct {
	detector      *SessionDetector
	reportService *service.ReportService
}

func NewReportGenerator(pool *pgxpool.Pool) *ReportGenerator {
	return &ReportGenerator{
		detector:      NewSessionDetector(pool),
		reportService: service.NewReportService(pool),
	}
}

// Run periodically generates daily reports for finalized sessions.
func (g *ReportGenerator) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run once immediately on startup (design spec 6.5 step 3)
	g.generate(ctx)

	for {
		select {
		case <-ticker.C:
			g.generate(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (g *ReportGenerator) generate(ctx context.Context) {
	dates, err := g.detector.GetFinalizedSessionDates(ctx)
	if err != nil {
		slog.Error("failed to get finalized session dates", slog.String("error", err.Error()))
		return
	}

	for _, date := range dates {
		if err := g.reportService.GenerateForDate(ctx, date); err != nil {
			slog.Error("failed to generate daily report",
				slog.String("date", date.Format("2006-01-02")),
				slog.String("error", err.Error()),
			)
		}
	}
}
