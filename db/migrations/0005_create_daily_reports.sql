-- +goose Up
CREATE TABLE daily_reports (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    report_date          DATE NOT NULL UNIQUE,
    session_count        INT NOT NULL DEFAULT 0,
    total_memos          INT NOT NULL DEFAULT 0,
    total_chars          INT NOT NULL DEFAULT 0,
    chars_per_minute     FLOAT,
    active_minutes       INT NOT NULL DEFAULT 0,
    hourly_distribution  JSONB NOT NULL DEFAULT '{}',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_daily_reports_date ON daily_reports (report_date DESC);

-- +goose Down
DROP TABLE IF EXISTS daily_reports;
