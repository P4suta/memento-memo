-- +goose Up
CREATE TABLE sessions (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    started_at  TIMESTAMPTZ NOT NULL,
    ended_at    TIMESTAMPTZ NOT NULL,
    date_label  DATE NOT NULL,
    memo_count  INT NOT NULL DEFAULT 0,
    total_chars INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_sessions_date_label ON sessions (date_label);
CREATE INDEX idx_sessions_started_at ON sessions (started_at DESC);

-- +goose Down
DROP TABLE IF EXISTS sessions;
