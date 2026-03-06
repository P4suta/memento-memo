package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/internal/repository"
)

type SessionService struct {
	pool    *pgxpool.Pool
	queries *repository.Queries
}

func NewSessionService(pool *pgxpool.Pool) *SessionService {
	return &SessionService{
		pool:    pool,
		queries: repository.New(pool),
	}
}

func (s *SessionService) List(ctx context.Context, from, to time.Time) ([]repository.Session, error) {
	sessions, err := s.queries.ListSessions(ctx, repository.ListSessionsParams{
		FromDate: ToPgDate(from),
		ToDate:   ToPgDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return sessions, nil
}

type HeatmapEntry struct {
	DateLabel  string `json:"date_label"`
	MemoCount  int    `json:"memo_count"`
	TotalChars int    `json:"total_chars"`
}

func (s *SessionService) Heatmap(ctx context.Context, from, to time.Time) ([]HeatmapEntry, error) {
	rows, err := s.queries.GetHeatmapData(ctx, repository.GetHeatmapDataParams{
		FromDate: ToPgDate(from),
		ToDate:   ToPgDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("get heatmap data: %w", err)
	}
	entries := make([]HeatmapEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, HeatmapEntry{
			DateLabel:  row.DateLabel.Time.Format("2006-01-02"),
			MemoCount:  int(row.MemoCount),
			TotalChars: int(row.TotalChars),
		})
	}
	return entries, nil
}

type DailySummary struct {
	Date     string               `json:"date"`
	Sessions []repository.Session `json:"sessions"`
	Memos    []repository.Memo    `json:"memos"`
}

func (s *SessionService) Daily(ctx context.Context, date time.Time) (*DailySummary, error) {
	pgDate := ToPgDate(date)
	sessions, err := s.queries.GetDailySummary(ctx, pgDate)
	if err != nil {
		return nil, fmt.Errorf("get daily summary: %w", err)
	}
	memos, err := s.queries.GetMemosByDateLabel(ctx, pgDate)
	if err != nil {
		return nil, fmt.Errorf("get memos by date: %w", err)
	}
	return &DailySummary{
		Date:     date.Format("2006-01-02"),
		Sessions: sessions,
		Memos:    memos,
	}, nil
}
