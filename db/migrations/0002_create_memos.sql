-- +goose Up
CREATE TABLE memos (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id   BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    content      TEXT NOT NULL,
    content_html TEXT NOT NULL DEFAULT '',
    metadata     JSONB NOT NULL DEFAULT '{}',
    pinned       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_memos_session_id ON memos (session_id);
CREATE INDEX idx_memos_created_at ON memos (created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_memos_pinned ON memos (pinned) WHERE pinned = TRUE AND deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS memos;
