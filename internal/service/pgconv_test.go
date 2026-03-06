package service

import (
	"testing"
	"time"
)

func TestToPgTimestamptz_RoundTrip(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	pg := ToPgTimestamptz(now)
	if !pg.Valid {
		t.Fatal("expected Valid=true")
	}
	got := FromPgTimestamptz(pg)
	if !got.Equal(now) {
		t.Errorf("round trip failed: got %v, want %v", got, now)
	}
}

func TestToPgDate(t *testing.T) {
	d := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	pg := ToPgDate(d)
	if !pg.Valid {
		t.Fatal("expected Valid=true")
	}
	if !pg.Time.Equal(d) {
		t.Errorf("date mismatch: got %v, want %v", pg.Time, d)
	}
}

func TestToPgText(t *testing.T) {
	pg := ToPgText("hello")
	if !pg.Valid {
		t.Fatal("expected Valid=true")
	}
	if pg.String != "hello" {
		t.Errorf("text mismatch: got %q, want %q", pg.String, "hello")
	}
}

func TestToPgFloat8_NonNil(t *testing.T) {
	v := 3.14
	pg := ToPgFloat8(&v)
	if !pg.Valid {
		t.Fatal("expected Valid=true")
	}
	if pg.Float64 != 3.14 {
		t.Errorf("float mismatch: got %f, want 3.14", pg.Float64)
	}
}

func TestToPgFloat8_Nil(t *testing.T) {
	pg := ToPgFloat8(nil)
	if pg.Valid {
		t.Fatal("expected Valid=false for nil input")
	}
}

func TestPgTimestamptzPtr_Valid(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	pg := ToPgTimestamptz(now)
	ptr := PgTimestamptzPtr(pg)
	if ptr == nil {
		t.Fatal("expected non-nil pointer")
	}
	if !ptr.Equal(now) {
		t.Errorf("time mismatch: got %v, want %v", *ptr, now)
	}
}

func TestPgTimestamptzPtr_Invalid(t *testing.T) {
	pg := ToPgTimestamptz(time.Time{})
	pg.Valid = false
	ptr := PgTimestamptzPtr(pg)
	if ptr != nil {
		t.Errorf("expected nil for invalid timestamptz, got %v", *ptr)
	}
}
