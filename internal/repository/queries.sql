-- name: GetActiveSession :one
-- Get the most recent session that could still be active (FOR UPDATE lock)
SELECT id, started_at, ended_at, date_label, memo_count, total_chars
FROM sessions
ORDER BY started_at DESC
LIMIT 1
FOR UPDATE;

-- name: CreateSession :one
INSERT INTO sessions (started_at, ended_at, date_label, memo_count, total_chars)
VALUES (@started_at, @started_at, @date_label, 0, 0)
RETURNING id, started_at, ended_at, date_label, memo_count, total_chars;

-- name: FinalizeSession :exec
UPDATE sessions SET ended_at = $2 WHERE id = $1;

-- name: UpdateSessionStats :exec
UPDATE sessions
SET memo_count = memo_count + 1,
    total_chars = total_chars + @chars::int,
    ended_at = @now
WHERE id = @id;

-- name: ListSessions :many
SELECT id, started_at, ended_at, date_label, memo_count, total_chars
FROM sessions
WHERE date_label >= @from_date AND date_label <= @to_date
ORDER BY started_at DESC;

-- name: GetHeatmapData :many
SELECT date_label, SUM(memo_count)::int AS memo_count, SUM(total_chars)::int AS total_chars
FROM sessions
WHERE date_label >= @from_date AND date_label <= @to_date
GROUP BY date_label
ORDER BY date_label;

-- name: GetDailySummary :many
SELECT id, started_at, ended_at, date_label, memo_count, total_chars
FROM sessions
WHERE date_label = $1
ORDER BY started_at;

-- name: CreateMemo :one
INSERT INTO memos (session_id, content, content_html)
VALUES ($1, $2, $3)
RETURNING id, session_id, content, content_html, metadata, pinned, created_at, updated_at, deleted_at;

-- name: GetMemo :one
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
WHERE m.id = $1;

-- name: ListMemos :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
WHERE m.deleted_at IS NULL
  AND (m.created_at < @cursor_created_at OR (m.created_at = @cursor_created_at AND m.id < @cursor_id))
ORDER BY m.created_at DESC, m.id DESC
LIMIT @page_limit;

-- name: ListMemosFirst :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
WHERE m.deleted_at IS NULL
ORDER BY m.created_at DESC, m.id DESC
LIMIT @page_limit;

-- name: ListMemosSince :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
WHERE m.deleted_at IS NULL AND m.created_at > @since
ORDER BY m.created_at DESC, m.id DESC
LIMIT @page_limit;

-- name: ListPinnedMemos :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
WHERE m.pinned = TRUE AND m.deleted_at IS NULL
ORDER BY m.created_at DESC;

-- name: UpdateMemo :one
UPDATE memos SET content = $2, content_html = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, session_id, content, content_html, metadata, pinned, created_at, updated_at, deleted_at;

-- name: SoftDeleteMemo :exec
UPDATE memos SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;

-- name: RestoreMemo :exec
UPDATE memos SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL;

-- name: PermanentDeleteMemo :exec
DELETE FROM memos WHERE id = $1;

-- name: TogglePin :one
UPDATE memos SET pinned = NOT pinned
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, session_id, content, content_html, metadata, pinned, created_at, updated_at, deleted_at;

-- name: ListDeletedMemos :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
WHERE m.deleted_at IS NOT NULL
ORDER BY m.deleted_at DESC
LIMIT @page_limit;

-- name: PurgeOldDeletedMemos :exec
DELETE FROM memos
WHERE deleted_at IS NOT NULL
  AND deleted_at < NOW() - INTERVAL '30 days';

-- name: SearchMemos :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
WHERE m.content ILIKE '%' || @query || '%'
  AND m.deleted_at IS NULL
  AND (m.created_at < @cursor_created_at OR (m.created_at = @cursor_created_at AND m.id < @cursor_id))
ORDER BY m.created_at DESC, m.id DESC
LIMIT @page_limit;

-- name: SearchMemosFirst :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
WHERE m.content ILIKE '%' || @query || '%'
  AND m.deleted_at IS NULL
ORDER BY m.created_at DESC, m.id DESC
LIMIT @page_limit;

-- name: UpsertTag :one
INSERT INTO tags (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name, created_at;

-- name: DeleteMemoTags :exec
DELETE FROM memo_tags WHERE memo_id = $1;

-- name: CreateMemoTag :exec
INSERT INTO memo_tags (memo_id, tag_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListTags :many
SELECT t.id, t.name, t.created_at, COUNT(mt.memo_id)::int AS memo_count
FROM tags t
JOIN memo_tags mt ON mt.tag_id = t.id
JOIN memos m ON m.id = mt.memo_id AND m.deleted_at IS NULL
GROUP BY t.id
ORDER BY memo_count DESC;

-- name: ListMemosByTag :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
JOIN memo_tags mt ON mt.memo_id = m.id
JOIN tags t ON t.id = mt.tag_id
WHERE t.name = @tag_name
  AND m.deleted_at IS NULL
  AND (m.created_at < @cursor_created_at OR (m.created_at = @cursor_created_at AND m.id < @cursor_id))
ORDER BY m.created_at DESC, m.id DESC
LIMIT @page_limit;

-- name: ListMemosByTagFirst :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
JOIN memo_tags mt ON mt.memo_id = m.id
JOIN tags t ON t.id = mt.tag_id
WHERE t.name = @tag_name
  AND m.deleted_at IS NULL
ORDER BY m.created_at DESC, m.id DESC
LIMIT @page_limit;

-- name: GetMemoTags :many
SELECT t.id, t.name, t.created_at
FROM tags t
JOIN memo_tags mt ON mt.tag_id = t.id
WHERE mt.memo_id = $1;

-- name: GetDailyReport :one
SELECT id, report_date, session_count, total_memos, total_chars,
       chars_per_minute, active_minutes, hourly_distribution, created_at
FROM daily_reports
WHERE report_date = $1;

-- name: UpsertDailyReport :one
INSERT INTO daily_reports (report_date, session_count, total_memos, total_chars, chars_per_minute, active_minutes, hourly_distribution)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (report_date) DO UPDATE SET
    session_count = EXCLUDED.session_count,
    total_memos = EXCLUDED.total_memos,
    total_chars = EXCLUDED.total_chars,
    chars_per_minute = EXCLUDED.chars_per_minute,
    active_minutes = EXCLUDED.active_minutes,
    hourly_distribution = EXCLUDED.hourly_distribution
RETURNING id, report_date, session_count, total_memos, total_chars,
          chars_per_minute, active_minutes, hourly_distribution, created_at;

-- name: GetStats :one
SELECT
    (SELECT COUNT(*) FROM memos WHERE deleted_at IS NULL)::int AS total_memos,
    (SELECT COUNT(*) FROM sessions)::int AS total_sessions,
    (SELECT COUNT(DISTINCT date_label) FROM sessions)::int AS active_days,
    (SELECT COALESCE(SUM(total_chars), 0) FROM sessions)::int AS total_chars;

-- name: GetFinalizedSessions :many
-- Sessions whose ended_at + gap threshold is before now (i.e., finalized)
SELECT id, started_at, ended_at, date_label, memo_count, total_chars
FROM sessions
WHERE ended_at < @threshold
ORDER BY date_label;

-- name: GetMemosBySession :many
SELECT id, session_id, content, content_html, metadata, pinned, created_at, updated_at, deleted_at
FROM memos
WHERE session_id = $1
ORDER BY created_at;

-- name: GetMemosByDateLabel :many
SELECT m.id, m.session_id, m.content, m.content_html, m.metadata, m.pinned,
       m.created_at, m.updated_at, m.deleted_at
FROM memos m
JOIN sessions s ON s.id = m.session_id
WHERE s.date_label = $1
ORDER BY m.created_at;
