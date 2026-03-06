package service

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/internal/repository"
)

const (
	SessionGapThreshold = 4 * time.Hour
	SessionMaxDuration  = 24 * time.Hour
	MaxMemoLength       = 10000
)

type MemoService struct {
	pool    *pgxpool.Pool
	queries *repository.Queries
}

func NewMemoService(pool *pgxpool.Pool) *MemoService {
	return &MemoService{
		pool:    pool,
		queries: repository.New(pool),
	}
}

type MemoWithTags struct {
	repository.Memo
	Tags []repository.Tag `json:"tags"`
}

func (s *MemoService) Create(ctx context.Context, content string) (*MemoWithTags, error) {
	if content == "" {
		return nil, ErrMemoEmpty
	}
	if utf8.RuneCountInString(content) > MaxMemoLength {
		return nil, ErrMemoTooLong
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// 1. Get active session (with row lock)
	session, err := qtx.GetActiveSession(ctx)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("get active session: %w", err)
	}

	now := time.Now()

	// 2. Session detection
	noSession := err == pgx.ErrNoRows
	gapExceeded := !noSession && now.Sub(session.EndedAt.Time) >= SessionGapThreshold
	durationExceeded := !noSession && now.Sub(session.StartedAt.Time) >= SessionMaxDuration

	if noSession || gapExceeded || durationExceeded {
		if !noSession {
			_ = qtx.FinalizeSession(ctx, repository.FinalizeSessionParams{
				ID:      session.ID,
				EndedAt: session.EndedAt,
			})
		}
		session, err = qtx.CreateSession(ctx, repository.CreateSessionParams{
			StartedAt: ToPgTimestamptz(now),
			DateLabel: ToPgDate(now),
		})
		if err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
	}

	// 3. Create memo
	html := RenderHTML(content)
	memo, err := qtx.CreateMemo(ctx, repository.CreateMemoParams{
		SessionID:   session.ID,
		Content:     content,
		ContentHtml: html,
	})
	if err != nil {
		return nil, fmt.Errorf("create memo: %w", err)
	}

	// 4. Update session stats
	_ = qtx.UpdateSessionStats(ctx, repository.UpdateSessionStatsParams{
		ID:    session.ID,
		Chars: int32(utf8.RuneCountInString(content)),
		Now:   ToPgTimestamptz(now),
	})

	// 5. Extract and save tags
	tags, err := s.syncTags(ctx, qtx, memo.ID, content)
	if err != nil {
		return nil, fmt.Errorf("sync tags: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &MemoWithTags{Memo: memo, Tags: tags}, nil
}

func (s *MemoService) Get(ctx context.Context, id int64) (*MemoWithTags, error) {
	memo, err := s.queries.GetMemo(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrMemoNotFound
		}
		return nil, fmt.Errorf("get memo: %w", err)
	}
	tags, err := s.queries.GetMemoTags(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get memo tags: %w", err)
	}
	return &MemoWithTags{Memo: memo, Tags: tags}, nil
}

func (s *MemoService) Update(ctx context.Context, id int64, content string) (*MemoWithTags, error) {
	if content == "" {
		return nil, ErrMemoEmpty
	}
	if utf8.RuneCountInString(content) > MaxMemoLength {
		return nil, ErrMemoTooLong
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	html := RenderHTML(content)
	memo, err := qtx.UpdateMemo(ctx, repository.UpdateMemoParams{
		ID:          id,
		Content:     content,
		ContentHtml: html,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrMemoNotFound
		}
		return nil, fmt.Errorf("update memo: %w", err)
	}

	tags, err := s.syncTags(ctx, qtx, id, content)
	if err != nil {
		return nil, fmt.Errorf("sync tags: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &MemoWithTags{Memo: memo, Tags: tags}, nil
}

func (s *MemoService) Delete(ctx context.Context, id int64) error {
	return s.queries.SoftDeleteMemo(ctx, id)
}

func (s *MemoService) Restore(ctx context.Context, id int64) error {
	return s.queries.RestoreMemo(ctx, id)
}

func (s *MemoService) PermanentDelete(ctx context.Context, id int64) error {
	return s.queries.PermanentDeleteMemo(ctx, id)
}

func (s *MemoService) TogglePin(ctx context.Context, id int64) (*MemoWithTags, error) {
	memo, err := s.queries.TogglePin(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrMemoNotFound
		}
		return nil, fmt.Errorf("toggle pin: %w", err)
	}
	tags, err := s.queries.GetMemoTags(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get memo tags: %w", err)
	}
	return &MemoWithTags{Memo: memo, Tags: tags}, nil
}

func (s *MemoService) syncTags(ctx context.Context, qtx *repository.Queries, memoID int64, content string) ([]repository.Tag, error) {
	_ = qtx.DeleteMemoTags(ctx, memoID)

	tagNames := ExtractTags(content)
	tags := make([]repository.Tag, 0, len(tagNames))
	for _, name := range tagNames {
		tag, err := qtx.UpsertTag(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("upsert tag %q: %w", name, err)
		}
		_ = qtx.CreateMemoTag(ctx, repository.CreateMemoTagParams{
			MemoID: memoID,
			TagID:  tag.ID,
		})
		tags = append(tags, tag)
	}
	return tags, nil
}
