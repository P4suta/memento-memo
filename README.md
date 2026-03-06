# Memento Memo

**セルフホスト型パーソナルメモアプリケーション**

Twitter風のタイムラインUIで、日々の思考・学習・アイデアを即座にメモとして記録する「お一人様SNS」。Docker Compose一発で起動し、ユーザー環境を一切汚さない完全自己完結型のアプリケーション。

---

## Demo

> TODO: スクリーンショット / GIF を追加

---

## Features

| 機能 | 説明 |
|------|------|
| **Twitter風タイムライン** | 投稿フォーム + 無限スクロールの時系列メモ一覧 |
| **Markdown対応** | GFM拡張（テーブル、タスクリスト、取り消し線、オートリンク）+ コードブロック |
| **リアルタイム更新** | PostgreSQL LISTEN/NOTIFY + WebSocket による即時反映 |
| **自動セッション検出** | 4時間のギャップ閾値で作業セッションを自動分割。設定不要 |
| **GitHub草ヒートマップ** | 年間の執筆活動をカレンダー形式で可視化 |
| **全文検索** | pg_trgm による日本語対応の部分一致検索 |
| **自動タグ抽出** | 本文中の `#タグ` を自動認識。Unicode対応で日本語タグも利用可能 |
| **日報自動生成** | セッション終了後にメモ数・文字数・字速・時間帯分布を自動集計 |
| **ソフトデリート** | ゴミ箱に30日保持後、自動で物理削除 |
| **ピン留め** | 重要なメモをタイムライン上部に固定 |
| **ゼロコンフィグ** | 環境変数のデフォルト値で即座に動作。設定ファイル不要 |

---

## Tech Stack

### Backend

| 技術 | 用途 | 選定理由 |
|------|------|----------|
| **Go 1.23** | APIサーバー | 単一バイナリ、低メモリ、高速起動。`net/http` の拡張パターンマッチング（Go 1.22+）により外部ルーター不要 |
| **PostgreSQL 17** | データストア | pg_trgm で形態素解析不要の日本語検索、LISTEN/NOTIFY でリアルタイム通知、JSONB で柔軟な拡張 |
| **pgx/v5** | DBドライバー | PostgreSQL専用ドライバー。database/sql を経由しない直接接続で最大パフォーマンス |
| **sqlc** | SQLコード生成 | SQLファーストのアプローチ。手書きSQLから型安全なGoコードを自動生成 |
| **goose** | マイグレーション | SQL駆動。embed.FS でバイナリに同梱し、起動時に自動実行 |
| **goldmark** + **bluemonday** | Markdown処理 | GFM拡張対応の変換 + HTMLサニタイズによるXSS防止 |
| **nhooyr.io/websocket** | WebSocket | 標準ライブラリスタイルのAPI。gorilla/websocket より軽量 |

### Frontend

| 技術 | 用途 | 選定理由 |
|------|------|----------|
| **SvelteKit** + **Svelte 5** | UIフレームワーク | Runes（`$state`, `$derived`, `$effect`）による直感的なリアクティビティ。外部状態管理ライブラリ不要 |
| **adapter-static** | ビルド | 静的SPAとしてビルドし、Goバイナリに embed。本番環境にNode.jsプロセス不要 |
| **TailwindCSS v4** | スタイリング | ユーティリティファーストCSS。v4のCSS-firstアーキテクチャで設定ファイル不要 |

### Infrastructure

| 技術 | 用途 |
|------|------|
| **Docker** + **Docker Compose** | コンテナオーケストレーション |
| **GitHub Actions** | CI/CD |
| **GitHub Container Registry** | コンテナイメージ配布 |

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Docker Compose Network               │
│                                                       │
│  ┌─────────────────────────────────────────────────┐  │
│  │              app container (~20MB)               │  │
│  │                                                  │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │  │
│  │  │ REST API │  │ WS Hub   │  │ Static Files │  │  │
│  │  │ :3000    │  │          │  │ (embed.FS)   │  │  │
│  │  └────┬─────┘  └────┬─────┘  └──────────────┘  │  │
│  │       │              │                           │  │
│  │  ┌────┴──────────────┴──────┐                   │  │
│  │  │      Service Layer       │                   │  │
│  │  │  (session, memo, tag,    │                   │  │
│  │  │   search, report)        │                   │  │
│  │  └────────────┬─────────────┘                   │  │
│  │               │                                  │  │
│  │  ┌────────────┴─────────────┐  ┌────────────┐  │  │
│  │  │   Repository (sqlc)      │  │ Workers    │  │  │
│  │  │                          │  │ - Session  │  │  │
│  │  └────────────┬─────────────┘  │ - Report   │  │  │
│  │               │                └──────┬─────┘  │  │
│  └───────────────┼───────────────────────┼────────┘  │
│                  │                       │            │
│  ┌───────────────┴───────────────────────┴────────┐  │
│  │           db container (PostgreSQL 17)          │  │
│  │                                                  │  │
│  │  pg_trgm │ LISTEN/NOTIFY │ Triggers │ GIN Index │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

本番環境で稼働するプロセスは**厳密に2つ**のみ。Goバイナリ（API + 静的配信 + WebSocket + バックグラウンドジョブ）と PostgreSQL。合計メモリフットプリントは **100-150MB** で NAS 級の省リソース環境でも快適に動作する。

> 詳細: [docs/architecture.md](docs/architecture.md)

---

## Quick Start

### Prerequisites

- Docker Desktop (または Docker Engine + Docker Compose)

### 起動

```bash
git clone https://github.com/sakashita/memento-memo.git
cd memento-memo
docker compose up -d
```

ブラウザで http://localhost:3000 にアクセス。

### 停止

```bash
docker compose down
```

### 完全削除（データを含むすべてを削除）

```bash
docker compose down -v
```

---

## Development

### 開発環境の起動

```bash
# 開発用コンテナを起動（ホットリロード対応）
docker compose -f docker-compose.dev.yml up -d

# Go API: http://localhost:3000 (air によるホットリロード)
# SvelteKit: http://localhost:5173 (Vite dev server)
```

### テスト実行

```bash
# Go ユニットテスト
docker run --rm -v $(pwd):/app -w /app golang:1.23-alpine \
  go test ./internal/service/...

# Go 統合テスト (testcontainers-go + PostgreSQL)
docker run --rm \
  -v $(pwd):/app \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e TESTCONTAINERS_RYUK_DISABLED=true \
  -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal \
  --add-host=host.docker.internal:host-gateway \
  -w /app golang:1.23-alpine \
  sh -c 'apk add --no-cache gcc musl-dev && go test -tags=integration -v ./internal/integration_test/...'

# フロントエンドテスト
docker run --rm -v $(pwd)/web:/app -w /app node:22-alpine npm test
```

### テスト構成

| カテゴリ | フレームワーク | テスト数 | 対象 |
|----------|---------------|---------|------|
| Go ユニット | `go test` | 30 | Cursor, Markdown, Tag, pgconv, Errors |
| Go 統合 | `testcontainers-go` | 19 | Memo CRUD, Session, Search, Tag, Pagination |
| Frontend | `vitest` | 12 | APIクライアント |
| **合計** | | **61** | |

> 詳細: [docs/development-guide.md](docs/development-guide.md)

---

## Project Structure

```
memento-memo/
├── cmd/memento/main.go          # エントリーポイント・起動シーケンス
├── internal/
│   ├── config/config.go         # 環境変数から設定読み取り
│   ├── server/
│   │   ├── server.go            # HTTPサーバー初期化・ルーティング
│   │   ├── middleware.go        # Recovery, Logger, SecurityHeaders, RateLimit
│   │   └── spa.go              # SPA フォールバックハンドラー
│   ├── handler/                 # HTTPハンドラー層
│   │   ├── memo.go             # メモ CRUD
│   │   ├── search.go           # 全文検索
│   │   ├── tag.go              # タグ一覧・絞り込み
│   │   ├── session.go          # セッション・カレンダー・ヒートマップ
│   │   ├── report.go           # 日報・統計
│   │   ├── health.go           # ヘルスチェック
│   │   ├── ws.go               # WebSocket アップグレード
│   │   └── errors.go           # 統一エラーレスポンス
│   ├── service/                 # ビジネスロジック層
│   │   ├── memo.go             # セッション検出 + メモ作成 (トランザクション制御)
│   │   ├── session.go          # セッション一覧・ヒートマップ・日別サマリー
│   │   ├── search.go           # カーソルページネーション付き検索
│   │   ├── report.go           # 日報生成・統計算出
│   │   ├── tag.go              # タグ抽出 (Unicode対応正規表現)
│   │   ├── markdown.go         # goldmark + bluemonday によるHTML変換
│   │   ├── cursor.go           # (created_at, id) 複合カーソル
│   │   ├── pgconv.go           # Go ⇔ pgtype 型変換ユーティリティ
│   │   └── errors.go           # AppError 型 + 定義済みエラー
│   ├── repository/              # データアクセス層 (sqlc自動生成)
│   │   ├── queries.sql         # 全SQLクエリ定義
│   │   ├── queries.sql.go      # sqlc生成コード
│   │   ├── models.go           # sqlc生成モデル
│   │   └── db.go               # DBTX インターフェース
│   ├── ws/                      # WebSocket リアルタイム通信
│   │   ├── hub.go              # 接続管理・ブロードキャスト
│   │   ├── client.go           # 個別クライアント (ReadPump/WritePump)
│   │   └── listener.go         # PostgreSQL LISTEN/NOTIFY リスナー
│   └── worker/                  # バックグラウンドワーカー
│       ├── session_detector.go  # ソフトデリート物理削除 (30日GC)
│       └── report_generator.go  # 日報自動生成 (5分間隔)
├── db/
│   ├── migrate.go              # goose マイグレーション実行 (embed.FS)
│   └── migrations/             # SQL マイグレーションファイル
│       ├── 0001_create_sessions.sql
│       ├── 0002_create_memos.sql
│       ├── 0003_create_tags.sql
│       ├── 0004_create_attachments.sql
│       ├── 0005_create_daily_reports.sql
│       └── 0006_add_triggers.sql
├── web/                         # SvelteKit フロントエンド
│   ├── embed.go / embed_dev.go # ビルドタグによる dev/prod 切り替え
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/            # APIクライアント + 型定義
│   │   │   ├── components/     # Svelte コンポーネント
│   │   │   └── stores/         # 状態管理 (Svelte 5 Runes)
│   │   └── routes/             # SvelteKit ルーティング
│   └── vitest.config.ts
├── Dockerfile                   # マルチステージビルド (Node → Go → Alpine)
├── Dockerfile.dev              # 開発用 (Go + Node 同梱)
├── docker-compose.yml          # 本番用
├── docker-compose.dev.yml      # 開発用 (ホットリロード)
├── sqlc.yaml                   # sqlc コード生成設定
└── .air.toml                   # Go ホットリロード設定
```

---

## Design Decisions

### なぜセッション自動検出か

カレンダーやヒートマップは「日」単位の集約を必要とするが、夜型ユーザーにとって0時境界は不自然。本アプリでは **4時間のギャップ閾値** と **24時間の継続上限** のデュアル条件で、ユーザーに一切の設定を要求せずにセッション境界を自動検出する。

```
23:00 メモ投稿 ─┐
23:30 メモ投稿  │ 同一セッション (date_label: 2026-03-06)
02:00 メモ投稿  │ ← 日付跨ぎだが4h未満なので継続
02:30 メモ投稿 ─┘
                    ← 4時間以上の空白（睡眠）
09:00 メモ投稿 ─┐ 新セッション (date_label: 2026-03-07)
09:15 メモ投稿 ─┘
```

### なぜカーソルページネーションか

オフセットベースのページネーション (`OFFSET 20`) は、新規メモ追加時にページ境界がズレて同じメモが重複表示される。本アプリでは `(created_at, id)` の複合カーソルにより、リアルタイム更新下でも正確なページングを保証する。

### なぜ sqlc か

ORMの暗黙的なクエリ生成は、パフォーマンスの予測可能性を損なう。sqlc は「SQLを書き、型安全なGoコードを得る」という SQLファーストのアプローチで、実行されるSQLが常に明示的。PostgreSQL固有機能（`GENERATED ALWAYS AS IDENTITY`、`gin_trgm_ops`、`FOR UPDATE`）も自然に利用できる。

### なぜフロントエンドを embed するか

本番環境でNode.jsプロセスを動かす必要がない。Go単一バイナリにSPAの静的アセットを `embed.FS` で内包し、同一オリジンで配信する。これにより CORS設定不要、デプロイが `docker compose up -d` の一行で完結、最終イメージサイズは約20MB。

---

## API Reference

> 詳細: [docs/api-reference.md](docs/api-reference.md)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/memos` | メモ作成 |
| `GET` | `/api/v1/memos` | メモ一覧 (cursor pagination) |
| `GET` | `/api/v1/memos/:id` | メモ詳細 |
| `PATCH` | `/api/v1/memos/:id` | メモ更新 |
| `DELETE` | `/api/v1/memos/:id` | メモ論理削除 |
| `POST` | `/api/v1/memos/:id/restore` | メモ復元 |
| `DELETE` | `/api/v1/memos/:id/permanent` | メモ物理削除 |
| `POST` | `/api/v1/memos/:id/pin` | ピン留め切り替え |
| `GET` | `/api/v1/search?q=...` | 全文検索 |
| `GET` | `/api/v1/tags` | タグ一覧 |
| `GET` | `/api/v1/tags/:name/memos` | タグ別メモ一覧 |
| `GET` | `/api/v1/sessions` | セッション一覧 |
| `GET` | `/api/v1/calendar/heatmap` | ヒートマップデータ |
| `GET` | `/api/v1/calendar/daily/:date` | 日別サマリー |
| `GET` | `/api/v1/reports/daily/:date` | 日報 |
| `GET` | `/api/v1/reports/stats` | 統計ダッシュボード |
| `GET` | `/api/v1/health` | ヘルスチェック |
| `GET` | `/api/v1/ws` | WebSocket接続 |

---

## Database Schema

> 詳細: [docs/database-design.md](docs/database-design.md)

```
sessions ──< memos ──< memo_tags >── tags
                  ──< attachments
daily_reports (sessions.date_label = report_date で結合)
```

5テーブル + 1中間テーブル。PostgreSQL固有機能（`GENERATED ALWAYS AS IDENTITY`、`TIMESTAMPTZ`、`JSONB`、`pg_trgm`、`LISTEN/NOTIFY`）を積極活用。

---

## Deployment

### バックアップ

```bash
# データベースバックアップ
docker compose exec db pg_dump -U memento -Fc memento > backup.dump

# リストア
docker compose exec -T db pg_restore -U memento -d memento --clean --if-exists < backup.dump
```

### 外部公開（任意）

本アプリは自身ではTLSを終端しない。外部ネットワークからアクセスする場合は、リバースプロキシまたはトンネルサービスを利用する。

```
# Caddyfile
memo.example.com {
    reverse_proxy app:3000
}
```

> 詳細: [docs/development-guide.md](docs/development-guide.md)

---

## Configuration

| 環境変数 | デフォルト | 説明 |
|----------|-----------|------|
| `DATABASE_URL` | (必須) | PostgreSQL接続文字列 |
| `PORT` | `3000` | HTTPサーバーポート |
| `LOG_LEVEL` | `info` | ログレベル (`debug`, `info`, `warn`, `error`) |
| `TZ` | `UTC` | タイムゾーン |

---

## License

MIT

---

## Author

坂下 康信 (Yasunobu Sakashita)
