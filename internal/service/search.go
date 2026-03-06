package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/internal/repository"
)

type SearchService struct {
	pool    *pgxpool.Pool
	queries *repository.Queries
}

func NewSearchService(pool *pgxpool.Pool) *SearchService {
	return &SearchService{
		pool:    pool,
		queries: repository.New(pool),
	}
}

func (s *SearchService) Search(ctx context.Context, query string, cursor *Cursor, limit int32) ([]repository.Memo, error) {
	if cursor != nil {
		memos, err := s.queries.SearchMemos(ctx, repository.SearchMemosParams{
			Query:           ToPgText(query),
			CursorCreatedAt: ToPgTimestamptz(cursor.CreatedAt),
			CursorID:        cursor.ID,
			PageLimit:       limit,
		})
		if err != nil {
			return nil, fmt.Errorf("search memos: %w", err)
		}
		return memos, nil
	}

	memos, err := s.queries.SearchMemosFirst(ctx, repository.SearchMemosFirstParams{
		Query:     ToPgText(query),
		PageLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search memos first: %w", err)
	}
	return memos, nil
}
