package service

import (
	"testing"
	"time"
)

func TestEncodeDecode_RoundTrip(t *testing.T) {
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	encoded := EncodeCursor(now, 42)
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decoded.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", decoded.CreatedAt, now)
	}
	if decoded.ID != 42 {
		t.Errorf("ID mismatch: got %d, want 42", decoded.ID)
	}
}

func TestDecodeCursor_InvalidBase64(t *testing.T) {
	_, err := DecodeCursor("!!!not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecodeCursor_InvalidJSON(t *testing.T) {
	// Valid base64 but not valid JSON
	_, err := DecodeCursor("bm90LWpzb24=") // "not-json"
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDecodeCursor_MissingFields(t *testing.T) {
	// Zero time (missing created_at)
	encoded := EncodeCursor(time.Time{}, 1)
	_, err := DecodeCursor(encoded)
	if err == nil {
		t.Fatal("expected error for zero time")
	}

	// Zero ID (missing id)
	encoded = EncodeCursor(time.Now(), 0)
	_, err = DecodeCursor(encoded)
	if err == nil {
		t.Fatal("expected error for zero ID")
	}
}
