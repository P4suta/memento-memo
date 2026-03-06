package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/internal/repository"
)

type ReportService struct {
	pool    *pgxpool.Pool
	queries *repository.Queries
}

func NewReportService(pool *pgxpool.Pool) *ReportService {
	return &ReportService{
		pool:    pool,
		queries: repository.New(pool),
	}
}

type DailyReport struct {
	repository.DailyReport
}

func (s *ReportService) GetDaily(ctx context.Context, date time.Time) (*DailyReport, error) {
	report, err := s.queries.GetDailyReport(ctx, ToPgDate(date))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get daily report: %w", err)
	}
	return &DailyReport{DailyReport: report}, nil
}

type Stats struct {
	TotalMemos    int `json:"total_memos"`
	TotalSessions int `json:"total_sessions"`
	ActiveDays    int `json:"active_days"`
	TotalChars    int `json:"total_chars"`
}

func (s *ReportService) GetStats(ctx context.Context) (*Stats, error) {
	row, err := s.queries.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return &Stats{
		TotalMemos:    int(row.TotalMemos),
		TotalSessions: int(row.TotalSessions),
		ActiveDays:    int(row.ActiveDays),
		TotalChars:    int(row.TotalChars),
	}, nil
}

// GenerateForDate generates or updates a daily report for the given date.
func (s *ReportService) GenerateForDate(ctx context.Context, date time.Time) error {
	pgDate := ToPgDate(date)
	sessions, err := s.queries.GetDailySummary(ctx, pgDate)
	if err != nil {
		return fmt.Errorf("get daily summary: %w", err)
	}
	if len(sessions) == 0 {
		return nil
	}

	var totalMemos, totalChars, activeMinutes int
	hourlyDist := make(map[string]int)

	for _, sess := range sessions {
		totalMemos += int(sess.MemoCount)
		totalChars += int(sess.TotalChars)

		duration := sess.EndedAt.Time.Sub(sess.StartedAt.Time)
		activeMinutes += int(duration.Minutes())

		memos, err := s.queries.GetMemosBySession(ctx, sess.ID)
		if err != nil {
			return fmt.Errorf("get memos by session: %w", err)
		}
		for _, m := range memos {
			hour := fmt.Sprintf("%d", m.CreatedAt.Time.Hour())
			hourlyDist[hour]++
		}
	}

	var cpm *float64
	if activeMinutes > 0 {
		v := float64(totalChars) / float64(activeMinutes)
		cpm = &v
	}

	hourlyJSON, err := json.Marshal(hourlyDist)
	if err != nil {
		return fmt.Errorf("marshal hourly dist: %w", err)
	}

	_, err = s.queries.UpsertDailyReport(ctx, repository.UpsertDailyReportParams{
		ReportDate:         pgDate,
		SessionCount:       int32(len(sessions)),
		TotalMemos:         int32(totalMemos),
		TotalChars:         int32(totalChars),
		CharsPerMinute:     ToPgFloat8(cpm),
		ActiveMinutes:      int32(activeMinutes),
		HourlyDistribution: json.RawMessage(hourlyJSON),
	})
	if err != nil {
		return fmt.Errorf("upsert daily report: %w", err)
	}

	return nil
}
