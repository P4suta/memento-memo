package service

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func ToPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func ToPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func ToPgFloat8(v *float64) pgtype.Float8 {
	if v == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *v, Valid: true}
}

func FromPgTimestamptz(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func PgTimestamptzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}
