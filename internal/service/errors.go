package service

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *AppError) Error() string { return e.Message }

var (
	ErrMemoNotFound = &AppError{Code: "MEMO_NOT_FOUND", Message: "指定されたメモは存在しません", Status: 404}
	ErrMemoTooLong  = &AppError{Code: "MEMO_TOO_LONG", Message: "メモは10,000文字以内で入力してください", Status: 400}
	ErrMemoEmpty    = &AppError{Code: "MEMO_EMPTY", Message: "メモ本文を入力してください", Status: 400}
	ErrMemoDeleted  = &AppError{Code: "MEMO_DELETED", Message: "削除済みのメモです", Status: 410}
	ErrTagNotFound  = &AppError{Code: "TAG_NOT_FOUND", Message: "指定されたタグは存在しません", Status: 404}
	ErrRateLimited  = &AppError{Code: "RATE_LIMITED", Message: "リクエスト制限を超過しました", Status: 429}
	ErrValidation   = &AppError{Code: "VALIDATION_ERROR", Message: "リクエストが不正です", Status: 400}
	ErrInternal     = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", Status: 500}
)
