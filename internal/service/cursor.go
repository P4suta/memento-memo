package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

func EncodeCursor(createdAt time.Time, id int64) string {
	c := Cursor{CreatedAt: createdAt, ID: id}
	data, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(data)
}

func DecodeCursor(s string) (*Cursor, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("invalid cursor data: %w", err)
	}
	if c.CreatedAt.IsZero() || c.ID == 0 {
		return nil, fmt.Errorf("invalid cursor: missing fields")
	}
	return &c, nil
}
