-- +goose Up
CREATE TABLE tags (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE memo_tags (
    memo_id BIGINT NOT NULL REFERENCES memos(id) ON DELETE CASCADE,
    tag_id  BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (memo_id, tag_id)
);

CREATE INDEX idx_memo_tags_tag_id ON memo_tags (tag_id);

-- +goose Down
DROP TABLE IF EXISTS memo_tags;
DROP TABLE IF EXISTS tags;
