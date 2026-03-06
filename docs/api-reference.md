# API Reference

Memento Memo REST API の仕様書。すべてのエンドポイントは `/api/v1` プレフィックスを持つ。

---

## 共通仕様

### リクエスト

- Content-Type: `application/json`
- リクエストボディ上限: 1MB

### レスポンス

- Content-Type: `application/json`
- 文字エンコーディング: UTF-8

### ページネーション

カーソルベースのページネーションを採用する。オフセットベース (`OFFSET N`) は新規データ追加時にページ境界がズレるため不採用。

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `cursor` | string | - | Base64エンコードされた `(created_at, id)` 複合カーソル |
| `limit` | integer | 20 | 取得件数 (最大100) |

**レスポンスフォーマット:**

```json
{
  "items": [ ... ],
  "next_cursor": "eyJjcmVhdGVkX2F0IjoiMjAyNi0wMy0wNlQwMjoxNTowMCswOTowMCIsImlkIjo0Mn0=",
  "has_more": true
}
```

`has_more` が `false` の場合、`next_cursor` は空文字列。

### エラーレスポンス

すべてのエラーは統一フォーマットで返却する。

```json
{
  "error": {
    "code": "MEMO_NOT_FOUND",
    "message": "指定されたメモは存在しません",
    "status": 404
  }
}
```

### エラーコード一覧

| コード | HTTP Status | 説明 |
|--------|-------------|------|
| `MEMO_NOT_FOUND` | 404 | 指定されたメモが存在しない |
| `MEMO_TOO_LONG` | 400 | メモ本文が10,000文字を超過 |
| `MEMO_EMPTY` | 400 | メモ本文が空 |
| `MEMO_DELETED` | 410 | 削除済みメモへの操作 |
| `TAG_NOT_FOUND` | 404 | 指定されたタグが存在しない |
| `VALIDATION_ERROR` | 400 | リクエストバリデーションエラー |
| `RATE_LIMITED` | 429 | レートリミット超過 (100 req/s) |
| `INTERNAL_ERROR` | 500 | サーバー内部エラー |

---

## Memos

### POST /api/v1/memos

メモを作成する。セッションの自動検出・タグの自動抽出を行う。

**Request Body:**

```json
{
  "content": "今日は #Go のcontext.Contextを深掘りした。"
}
```

| フィールド | 型 | 必須 | 制約 |
|-----------|-----|------|------|
| `content` | string | Yes | 1〜10,000文字 |

**Response: 201 Created**

```json
{
  "id": 42,
  "session_id": 7,
  "content": "今日は #Go のcontext.Contextを深掘りした。",
  "content_html": "<p>今日は #Go のcontext.Contextを深掘りした。</p>\n",
  "tags": [
    { "id": 1, "name": "Go", "created_at": "2026-03-06T02:15:00+09:00" }
  ],
  "pinned": false,
  "created_at": "2026-03-06T02:15:00+09:00",
  "updated_at": "2026-03-06T02:15:00+09:00"
}
```

**処理フロー:**
1. バリデーション (空文字/文字数上限)
2. アクティブセッション判定 (4h ギャップ / 24h 上限)
3. Markdown → HTML 変換 (goldmark + bluemonday)
4. メモをDBに保存
5. セッション統計更新 (memo_count, total_chars)
6. タグ抽出・同期 (`#タグ` を正規表現で抽出)
7. PostgreSQL トリガーが `memo_changes` を NOTIFY → WebSocket 配信

---

### GET /api/v1/memos

メモ一覧を取得する。カーソルページネーション対応。

**Query Parameters:**

| パラメータ | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `cursor` | string | - | ページネーションカーソル |
| `limit` | integer | 20 | 取得件数 |
| `since` | string (ISO 8601) | - | 指定タイムスタンプ以降のメモのみ取得 |

`since` パラメータは WebSocket 再接続時の差分取得に使用する。`cursor` と `since` が同時に指定された場合、`since` を優先する。

**Response: 200 OK**

```json
{
  "items": [
    {
      "id": 42,
      "session_id": 7,
      "content": "...",
      "content_html": "...",
      "tags": [],
      "pinned": false,
      "created_at": "2026-03-06T02:15:00+09:00",
      "updated_at": "2026-03-06T02:15:00+09:00"
    }
  ],
  "next_cursor": "eyJjcm...",
  "has_more": true
}
```

---

### GET /api/v1/memos/:id

メモ詳細を取得する。

**Response: 200 OK**

メモ作成レスポンスと同一フォーマット。

**Error:** `MEMO_NOT_FOUND` (404)

---

### PATCH /api/v1/memos/:id

メモを更新する。タグの再抽出・同期を行う。

**Request Body:**

```json
{
  "content": "更新された内容 #新しいタグ"
}
```

**Response: 200 OK**

更新後のメモオブジェクト。`content_html` と `tags` が再生成される。

**処理フロー:**
1. Markdown → HTML 再変換
2. 既存タグ関連を全削除 → 再抽出したタグで再挿入 (delete-and-recreate)
3. `updated_at` トリガーが自動更新

---

### DELETE /api/v1/memos/:id

メモを論理削除する（ソフトデリート）。`deleted_at` が設定され、タイムラインから非表示になるがDBには残る。30日後にバックグラウンドワーカーが物理削除する。

**Response: 204 No Content**

---

### POST /api/v1/memos/:id/restore

論理削除されたメモを復元する。

**Response: 204 No Content**

---

### DELETE /api/v1/memos/:id/permanent

メモを即座に物理削除する。この操作は取り消せない。`memo_tags`、`attachments` は `ON DELETE CASCADE` で連鎖削除される。

**Response: 204 No Content**

---

### POST /api/v1/memos/:id/pin

メモのピン留め状態を切り替える。

**Response: 200 OK**

ピン状態が反転したメモオブジェクト。

---

## Search

### GET /api/v1/search

メモの全文検索を行う。PostgreSQL の `pg_trgm` 拡張による `ILIKE` 部分一致検索。日本語対応。

**Query Parameters:**

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `q` | string | Yes | 検索クエリ |
| `cursor` | string | No | ページネーションカーソル |
| `limit` | integer | No | 取得件数 (デフォルト: 20) |

**Response: 200 OK**

ページネーション付きメモ一覧。

> **技術的補足:** `pg_trgm` はテキストをトライグラム（3文字の部分文字列）に分解してGINインデックスで検索する。形態素解析不要で日本語の部分一致検索が動作する。`ILIKE` は `gin_trgm_ops` インデックスで高速化され、メモ数10万件まで <50ms で応答する。

---

## Tags

### GET /api/v1/tags

タグ一覧を使用数降順で取得する。

**Response: 200 OK**

```json
{
  "items": [
    { "id": 1, "name": "Go", "memo_count": 15, "created_at": "2026-03-01T10:00:00+09:00" },
    { "id": 2, "name": "Docker", "memo_count": 8, "created_at": "2026-03-02T14:00:00+09:00" }
  ]
}
```

---

### GET /api/v1/tags/:name/memos

指定タグに紐づくメモ一覧を取得する。カーソルページネーション対応。

**Path Parameters:**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `name` | string | タグ名（URLエンコード必須） |

**Response: 200 OK**

ページネーション付きメモ一覧。

---

## Sessions / Calendar

### GET /api/v1/sessions

セッション一覧を取得する。

**Query Parameters:**

| パラメータ | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `from` | string (YYYY-MM-DD) | 30日前 | 開始日 |
| `to` | string (YYYY-MM-DD) | 今日 | 終了日 |

**Response: 200 OK**

```json
{
  "items": [
    {
      "id": 7,
      "started_at": "2026-03-06T00:30:00+09:00",
      "ended_at": "2026-03-06T03:15:00+09:00",
      "date_label": "2026-03-06",
      "memo_count": 12,
      "total_chars": 3500
    }
  ]
}
```

---

### GET /api/v1/calendar/heatmap

GitHub草スタイルのヒートマップデータを取得する。

**Query Parameters:**

| パラメータ | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `year` | integer | 今年 | 対象年 |

**Response: 200 OK**

```json
{
  "items": [
    { "date_label": "2026-03-06", "memo_count": 12, "total_chars": 3500 },
    { "date_label": "2026-03-07", "memo_count": 5, "total_chars": 1200 }
  ]
}
```

---

### GET /api/v1/calendar/daily/:date

日別の詳細サマリーを取得する。

**Path Parameters:**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `date` | string (YYYY-MM-DD) | 対象日 |

**Response: 200 OK**

```json
{
  "date": "2026-03-06",
  "sessions": [ ... ],
  "memos": [ ... ]
}
```

---

## Reports

### GET /api/v1/reports/daily/:date

日報を取得する。バックグラウンドワーカーが5分間隔で自動生成する。

**Response: 200 OK**

```json
{
  "id": 1,
  "report_date": "2026-03-06",
  "session_count": 2,
  "total_memos": 15,
  "total_chars": 4200,
  "chars_per_minute": 28.0,
  "active_minutes": 150,
  "hourly_distribution": {
    "0": 3, "1": 5, "2": 4, "9": 2, "10": 1
  },
  "created_at": "2026-03-07T00:05:00+09:00"
}
```

---

### GET /api/v1/reports/stats

全体統計を取得する。

**Response: 200 OK**

```json
{
  "total_memos": 1234,
  "total_sessions": 89,
  "active_days": 45,
  "total_chars": 350000
}
```

---

## Health

### GET /api/v1/health

ヘルスチェック。Docker Compose のヘルスチェックやロードバランサーのバックエンドチェックに使用。

**Response: 200 OK**

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "72h15m30s",
  "database": "connected"
}
```

---

## WebSocket

### GET /api/v1/ws

WebSocket接続を確立する。PostgreSQL の `LISTEN/NOTIFY` と連携し、メモの変更をリアルタイムに配信する。

**接続:** 標準の WebSocket アップグレード。

**メッセージフォーマット:**

```json
{
  "type": "memo.created | memo.updated | memo.deleted | ping | pong",
  "payload": { "memo_id": 42, "session_id": 7 },
  "timestamp": "2026-03-06T02:15:00+09:00"
}
```

**メッセージ型:**

| type | 方向 | 説明 |
|------|------|------|
| `memo.created` | Server → Client | メモが作成された |
| `memo.updated` | Server → Client | メモが更新された |
| `memo.deleted` | Server → Client | メモが削除された |
| `ping` | Client → Server | 生存確認 (30秒間隔) |
| `pong` | Server → Client | ping への応答 |

**再接続戦略:**
- 初回: 即座 (0ms)
- 2回目以降: Exponential Backoff (1s, 2s, 4s, 8s, ... 最大30s)
- タブ非アクティブ時は再接続を一時停止
- 再接続成功時に `since` パラメータで差分取得
