-- +goose Up
-- +goose StatementBegin

-- pg_trgm extension for trigram-based search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Trigram index for full-text search
CREATE INDEX idx_memos_content_trgm ON memos USING GIN (content gin_trgm_ops);

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_memos_updated_at
    BEFORE UPDATE ON memos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Real-time notification trigger
CREATE OR REPLACE FUNCTION notify_memo_change()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('memo_changes', json_build_object(
        'action', TG_OP,
        'memo_id', COALESCE(NEW.id, OLD.id),
        'session_id', COALESCE(NEW.session_id, OLD.session_id)
    )::text);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_memo_notify
    AFTER INSERT OR UPDATE OR DELETE ON memos
    FOR EACH ROW
    EXECUTE FUNCTION notify_memo_change();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_memo_notify ON memos;
DROP FUNCTION IF EXISTS notify_memo_change();
DROP TRIGGER IF EXISTS trg_memos_updated_at ON memos;
DROP FUNCTION IF EXISTS update_updated_at();
DROP INDEX IF EXISTS idx_memos_content_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd
