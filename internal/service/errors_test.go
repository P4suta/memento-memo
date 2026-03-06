package service

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := &AppError{Code: "TEST", Message: "test message", Status: 400}
	if err.Error() != "test message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test message")
	}
}

func TestAppError_ErrorsAs(t *testing.T) {
	var appErr *AppError
	err := error(ErrMemoNotFound)
	if !errors.As(err, &appErr) {
		t.Fatal("errors.As should match *AppError")
	}
	if appErr.Code != "MEMO_NOT_FOUND" {
		t.Errorf("Code = %q, want MEMO_NOT_FOUND", appErr.Code)
	}
	if appErr.Status != 404 {
		t.Errorf("Status = %d, want 404", appErr.Status)
	}
}

func TestAppError_PredefinedErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    *AppError
		status int
	}{
		{"MemoNotFound", ErrMemoNotFound, 404},
		{"MemoTooLong", ErrMemoTooLong, 400},
		{"MemoEmpty", ErrMemoEmpty, 400},
		{"MemoDeleted", ErrMemoDeleted, 410},
		{"TagNotFound", ErrTagNotFound, 404},
		{"RateLimited", ErrRateLimited, 429},
		{"Validation", ErrValidation, 400},
		{"Internal", ErrInternal, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Status != tt.status {
				t.Errorf("Status = %d, want %d", tt.err.Status, tt.status)
			}
			if tt.err.Error() == "" {
				t.Error("Error() should not be empty")
			}
			if tt.err.Code == "" {
				t.Error("Code should not be empty")
			}
		})
	}
}
