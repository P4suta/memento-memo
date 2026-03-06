# Development Guide

Memento Memo の開発環境構築からテスト実行、コード生成、デプロイまでのワークフローを解説する。

---

## 前提条件

| ツール | バージョン | 用途 |
|--------|-----------|------|
| Docker Desktop | 最新版 | コンテナ実行環境 |
| WSL2 (Ubuntu 24.04) | - | 開発ホスト |
| Go | 1.23+ | バックエンド開発 |
| Node.js | 22+ | フロントエンド開発 |
| Git | 2.x | バージョン管理 |

> **Note:** Go と Node.js は `Dockerfile.dev` 内に同梱されているため、WSL ホストへのインストールは必須ではない。Docker Compose 経由の開発ワークフローのみを使用する場合はスキップ可能。

---

## ディレクトリ構造

```
memento-memo/
├── cmd/
│   └── memento/
│       └── main.go              # エントリーポイント
├── db/
│   ├── migrations/              # goose SQLマイグレーション
│   │   ├── 0001_create_sessions.sql
│   │   ├── 0002_create_memos.sql
│   │   ├── 0003_create_tags.sql
│   │   ├── 0004_create_attachments.sql
│   │   ├── 0005_create_daily_reports.sql
│   │   └── 0006_add_triggers.sql
│   └── migrate.go               # embed.FS + goose ランナー
├── internal/
│   ├── config/
│   │   └── config.go            # 環境変数読み取り
│   ├── handler/                 # HTTPハンドラー (REST API)
│   ├── integration_test/        # 統合テスト (testcontainers-go)
│   ├── repository/              # sqlc 自動生成コード
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── queries.sql          # SQL クエリ定義
│   │   └── queries.sql.go       # 自動生成 (手動編集禁止)
│   ├── server/                  # HTTPサーバー, ミドルウェア, ルーティング
│   ├── service/                 # ビジネスロジック
│   │   ├── cursor.go            # カーソルページネーション
│   │   ├── errors.go            # アプリケーションエラー
│   │   ├── markdown.go          # Markdown → HTML 変換
│   │   ├── memo.go              # メモ CRUD + セッション検出
│   │   ├── pgconv.go            # Go ⇔ pgtype 変換
│   │   ├── report.go            # 日報生成
│   │   ├── search.go            # 全文検索
│   │   ├── session.go           # セッション管理
│   │   └── tag.go               # タグ抽出
│   ├── worker/                  # バックグラウンドワーカー
│   └── ws/                      # WebSocket Hub + LISTEN/NOTIFY
├── web/                         # SvelteKit フロントエンド
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/             # APIクライアント + 型定義
│   │   │   ├── components/      # Svelte コンポーネント
│   │   │   └── stores/          # Svelte 5 Runes 状態管理
│   │   └── routes/              # SvelteKit ルート
│   ├── embed.go                 # 本番用 embed.FS (ビルドタグ !dev)
│   ├── embed_dev.go             # 開発用 noop (ビルドタグ dev)
│   ├── package.json
│   ├── vite.config.ts
│   └── vitest.config.ts
├── docker-compose.yml           # 本番構成
├── docker-compose.dev.yml       # 開発構成
├── Dockerfile                   # マルチステージ本番ビルド
├── Dockerfile.dev               # 開発用 (Go + Node.js + ホットリロード)
├── go.mod
├── go.sum
└── sqlc.yaml                    # sqlc 設定
```

---

## 環境構築

### Docker Compose 開発環境 (推奨)

最も簡単な開発方法。Go と Node.js のインストールが不要。

```bash
# 1. リポジトリをクローン
git clone https://github.com/sakashita/memento-memo.git
cd memento-memo

# 2. 開発コンテナを起動
docker compose -f docker-compose.dev.yml up -d

# 3. ログを確認
docker compose -f docker-compose.dev.yml logs -f dev
```

開発コンテナが起動すると:
- **Go API サーバー** (port 3000): `air` によるホットリロード
- **Vite 開発サーバー** (port 5173): HMR 有効
- **PostgreSQL** (port 5432): ホストから直接接続可能

フロントエンドは `http://localhost:5173` でアクセスし、API リクエストは Vite のプロキシ設定により `http://localhost:3000` に転送される。

```typescript
// web/vite.config.ts - プロキシ設定
server: {
    proxy: {
        '/api': {
            target: 'http://localhost:3000',
            changeOrigin: true
        }
    }
}
```

### ネイティブ開発環境

WSL ホストに直接 Go と Node.js をインストールして開発する場合。

```bash
# Go 1.23 インストール
wget https://go.dev/dl/go1.23.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Node.js 22 インストール
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt-get install -y nodejs

# 開発ツールのインストール
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/air-verse/air@latest

# PostgreSQL を Docker で起動
docker run -d --name memento-db \
  -e POSTGRES_USER=memento \
  -e POSTGRES_PASSWORD=memento \
  -e POSTGRES_DB=memento \
  -p 5432:5432 \
  postgres:17-alpine \
  -c shared_buffers=128MB

# Go サーバー起動 (開発モード)
export DATABASE_URL="postgres://memento:memento@localhost:5432/memento?sslmode=disable"
air -c .air.toml

# フロントエンド起動 (別ターミナル)
cd web
npm install
npm run dev -- --host 0.0.0.0 --port 5173
```

---

## 環境変数

| 変数名 | 必須 | デフォルト | 説明 |
|--------|------|-----------|------|
| `DATABASE_URL` | Yes | - | PostgreSQL 接続文字列 |
| `PORT` | No | `3000` | HTTP サーバーポート |
| `LOG_LEVEL` | No | `info` | ログレベル (`debug`, `info`, `warn`, `error`) |
| `TZ` | No | `UTC` | タイムゾーン |

```bash
# 例: 開発用の設定
export DATABASE_URL="postgres://memento:memento@db:5432/memento?sslmode=disable"
export LOG_LEVEL=debug
export TZ=Asia/Tokyo
```

---

## コード生成ワークフロー

### sqlc: SQL → Go コード生成

SQL クエリを変更したら `sqlc generate` で Go コードを再生成する。

```bash
# クエリファイルを編集
vim internal/repository/queries.sql

# コード生成 (開発コンテナ内)
docker compose -f docker-compose.dev.yml exec dev sqlc generate

# またはネイティブ環境
sqlc generate
```

**設定ファイル (`sqlc.yaml`):**

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/repository/queries.sql"
    schema: "db/migrations"
    gen:
      go:
        package: "repository"
        out: "internal/repository"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_empty_slices: true
        overrides:
          - db_type: "jsonb"
            go_type:
              import: "encoding/json"
              type: "RawMessage"
```

**生成対象ファイル:**
- `internal/repository/db.go` - DBTX インターフェース
- `internal/repository/models.go` - テーブルに対応する Go 構造体
- `internal/repository/queries.sql.go` - クエリ関数

> **重要:** `queries.sql.go` と `models.go` は自動生成ファイルであり、手動で編集してはならない。変更は `queries.sql` に対して行い、`sqlc generate` で再生成する。

### goose: データベースマイグレーション

マイグレーションファイルは `db/migrations/` に配置し、`embed.FS` で Go バイナリに同梱される。アプリケーション起動時に自動実行されるため、通常は手動実行する必要はない。

```bash
# 新しいマイグレーション作成
goose -dir db/migrations create add_new_column sql

# 手動でマイグレーション実行
goose -dir db/migrations postgres "$DATABASE_URL" up

# マイグレーション状態確認
goose -dir db/migrations postgres "$DATABASE_URL" status

# ロールバック (1つ戻す)
goose -dir db/migrations postgres "$DATABASE_URL" down
```

**マイグレーションファイルの規約:**

```sql
-- +goose Up
CREATE TABLE example (...);

-- +goose Down
DROP TABLE IF EXISTS example;
```

すべてのマイグレーションには `-- +goose Up` と `-- +goose Down` の両方を定義する。

---

## ビルド

### 本番ビルド (Docker)

マルチステージビルドにより、最小限の実行イメージ (~20MB) を生成する。

```bash
# イメージビルド
docker compose build

# バージョン情報を埋め込む場合
docker build \
  --build-arg BUILD_VERSION=1.0.0 \
  --build-arg BUILD_COMMIT=$(git rev-parse --short HEAD) \
  -t memento-memo .
```

**ビルドステージ:**

| ステージ | ベースイメージ | 目的 |
|---------|--------------|------|
| Stage 1: `frontend` | `node:22-alpine` | SvelteKit ビルド (`npm ci` + `npm run build`) |
| Stage 2: `backend` | `golang:1.23-alpine` | Go バイナリビルド (フロントエンドを `embed.FS` で内包) |
| Stage 3: runtime | `alpine:3.20` | 実行イメージ (バイナリ + CA証明書 + tzdata) |

最終イメージは非 root ユーザー `memento` で実行される。

### 開発ビルド (ビルドタグ)

フロントエンドの配信方法はビルドタグで切り替わる:

```go
// web/embed.go (本番: ビルドタグ !dev)
//go:build !dev

//go:embed dist/*
var assets embed.FS
```

```go
// web/embed_dev.go (開発: ビルドタグ dev)
//go:build dev

// embed なし。Vite dev server にプロキシ
```

開発時は `-tags dev` をつけてビルドすることで、`web/dist/` ディレクトリが不要になる:

```bash
go build -tags dev -o memento ./cmd/memento
```

---

## テスト

### テストスイート概要

| スイート | テスト数 | 実行環境 | 説明 |
|---------|---------|---------|------|
| Go ユニットテスト | 30 | Docker / ネイティブ | DB不要。純粋なロジックテスト |
| Go 統合テスト | 19 | Docker-in-Docker | testcontainers-go + PostgreSQL |
| フロントエンドテスト | 12 | Docker / ネイティブ | vitest + jsdom |

### Go ユニットテスト

DB に依存しない純粋なロジックをテストする。

```bash
# 全ユニットテスト実行
go test ./internal/service/...

# Docker 経由
docker run --rm -v ~/memento-memo:/app -w /app golang:1.23-alpine \
  go test ./internal/service/...

# 特定テストのみ
go test -run TestExtractTags ./internal/service/...

# 詳細出力
go test -v ./internal/service/...
```

**テストファイル:**

| ファイル | テスト対象 | テスト数 |
|---------|-----------|---------|
| `cursor_test.go` | `EncodeCursor` / `DecodeCursor` | 4 |
| `markdown_test.go` | `RenderHTML` (goldmark + bluemonday) | 9 |
| `tag_test.go` | `ExtractTags` (Unicode 正規表現) | 7 |
| `pgconv_test.go` | pgtype 変換関数群 | 7 |
| `errors_test.go` | `AppError` エラー型 | 3 |

### Go 統合テスト

testcontainers-go を使用して実際の PostgreSQL コンテナを起動し、エンドツーエンドのデータフローをテストする。

```bash
# Docker-in-Docker で統合テスト実行
docker run --rm \
  -v ~/memento-memo:/app \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e TESTCONTAINERS_RYUK_DISABLED=true \
  -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal \
  --add-host=host.docker.internal:host-gateway \
  -w /app golang:1.23-alpine \
  go test -tags=integration -v ./internal/integration_test/...
```

**ビルドタグ:** `//go:build integration` — 通常の `go test` では実行されない。`-tags=integration` フラグが必要。

**テストインフラ (`testmain_test.go`):**

```go
func TestMain(m *testing.M) {
    // 1. PostgreSQL 17 Alpine コンテナを起動
    // 2. goose マイグレーションを実行
    // 3. pgxpool を作成して testPool に格納
    // 4. テスト実行
    // 5. コンテナ終了
}
```

各テスト関数の先頭で `cleanDB(t)` を呼び出し、`TRUNCATE ... RESTART IDENTITY CASCADE` で全テーブルをクリアする。テスト間の状態汚染を防止。

**テストファイル:**

| ファイル | テスト内容 | テスト数 |
|---------|-----------|---------|
| `memo_test.go` | Create→Get、Update、SoftDelete→Restore、PermanentDelete、TogglePin、バリデーションエラー | 6 |
| `session_test.go` | 新規セッション作成、連続投稿で同一セッション、4h ギャップで新規、24h 超過で新規、統計更新 | 5 |
| `search_test.go` | 全文検索ヒット、部分一致 (pg_trgm)、結果なし | 3 |
| `tag_test.go` | タグ付き作成→一覧反映、更新時タグ同期、タグ別絞り込み | 3 |
| `pagination_test.go` | limit+cursor 連続取得、since パラメータ | 2 |

**Docker-in-Docker 環境変数の解説:**

| 変数 | 値 | 理由 |
|------|-----|------|
| `TESTCONTAINERS_RYUK_DISABLED` | `true` | DinD 環境で Ryuk (リソースクリーナー) が通信できないため無効化 |
| `TESTCONTAINERS_HOST_OVERRIDE` | `host.docker.internal` | テストコードから PostgreSQL コンテナのマッピングポートへ到達するため |

### フロントエンドテスト

vitest + jsdom 環境で API クライアントをテストする。

```bash
# テスト実行
cd web
npm test

# Docker 経由
docker run --rm -v ~/memento-memo/web:/app -w /app node:22-alpine npm test

# ウォッチモード (開発中)
cd web
npx vitest

# カバレッジ
npx vitest run --coverage
```

**設定 (`web/vitest.config.ts`):**

```typescript
export default defineConfig({
    plugins: [sveltekit()],
    test: {
        environment: 'jsdom',
        include: ['src/**/*.test.ts']
    }
});
```

**テストファイル:**

| ファイル | テスト対象 | テスト数 |
|---------|-----------|---------|
| `client.test.ts` | 全 API メソッド (fetch モック)、エラーハンドリング | 12 |

---

## 起動シーケンス

アプリケーション起動時の処理順序 (`cmd/memento/main.go`):

```
1. 環境変数読み込み (config.MustLoad)
2. JSON 構造化ログ初期化 (slog)
3. DB 接続 (リトライ付き: 最大10回, 線形バックオフ)
4. マイグレーション自動実行 (goose embed.FS)
5. バックグラウンドワーカー起動
   ├── ReportGenerator (5分間隔で日報生成)
   └── SessionDetector (セッション終了検出)
6. WebSocket Hub + LISTEN/NOTIFY リスナー起動
7. HTTP サーバー起動 (port 3000)
8. SIGTERM/SIGINT 待機
9. Graceful Shutdown (30秒タイムアウト)
```

**DB 接続リトライ:**

```go
// 線形バックオフ: 1s, 2s, 3s, ..., 10s
for attempt := 0; attempt < 10; attempt++ {
    pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
    if err == nil {
        if pingErr := pool.Ping(ctx); pingErr == nil {
            return pool, nil
        }
    }
    time.Sleep(time.Duration(attempt+1) * time.Second)
}
```

**コネクションプール設定:**

| パラメータ | 値 | 説明 |
|-----------|-----|------|
| `MaxConns` | 10 | 最大接続数 |
| `MinConns` | 2 | 最小アイドル接続数 |
| `MaxConnLifetime` | 1h | 接続の最大生存期間 |
| `MaxConnIdleTime` | 30m | アイドル接続のタイムアウト |
| `HealthCheckPeriod` | 1m | ヘルスチェック間隔 |

---

## 本番デプロイ

### Docker Compose 起動

```bash
# 起動
docker compose up -d

# ステータス確認
docker compose ps

# ログ確認
docker compose logs -f app

# 停止
docker compose down

# データも含めて完全削除
docker compose down -v
```

### ヘルスチェック

```bash
# アプリケーション
curl http://localhost:3000/api/v1/health

# レスポンス
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "72h15m30s",
  "database": "connected"
}
```

### リソース制限

`docker-compose.yml` で設定済み:

| コンテナ | メモリ上限 | 想定使用量 |
|---------|----------|-----------|
| `app` | 128MB | 20-50MB |
| `db` | 384MB | 50-100MB |

### ログローテーション

両コンテナとも `json-file` ドライバーで最大 10MB × 3 ファイルに制限:

```yaml
logging:
  driver: json-file
  options:
    max-size: "10m"
    max-file: "3"
```

### データ永続化

| ボリューム | マウント先 | 内容 |
|-----------|----------|------|
| `pgdata` | `/var/lib/postgresql/data` | PostgreSQL データ |
| `attachments` | `/data/attachments` | 添付ファイル (Phase 2) |

### バックアップ

```bash
# PostgreSQL データのバックアップ
docker compose exec db pg_dump -U memento memento > backup_$(date +%Y%m%d).sql

# リストア
docker compose exec -T db psql -U memento memento < backup_20260306.sql
```

---

## 開発ガイドライン

### レイヤー間の依存ルール

```
handler → service → repository → PostgreSQL
```

- `handler` は `service` のみを呼ぶ。SQL を直接書かない
- `service` は `repository` のみを呼ぶ。HTTP レスポンスを構築しない
- `repository` は sqlc 自動生成コード。手動編集しない

### 新しいエンドポイントの追加手順

1. **SQL クエリ定義** — `internal/repository/queries.sql` にクエリを追加
2. **コード生成** — `sqlc generate` で Go コードを再生成
3. **サービス層** — `internal/service/` にビジネスロジックを実装
4. **ハンドラー層** — `internal/handler/` に HTTP ハンドラーを実装
5. **ルーティング** — `internal/server/` でルートを登録
6. **テスト** — ユニットテスト + 統合テストを追加

### エラーハンドリング

アプリケーションエラーは `service.AppError` 型を使用:

```go
// 定義済みエラー (internal/service/errors.go)
ErrMemoNotFound  = &AppError{Code: "MEMO_NOT_FOUND", Message: "...", Status: 404}
ErrMemoTooLong   = &AppError{Code: "MEMO_TOO_LONG", Message: "...", Status: 400}
ErrMemoEmpty     = &AppError{Code: "MEMO_EMPTY", Message: "...", Status: 400}
ErrMemoDeleted   = &AppError{Code: "MEMO_DELETED", Message: "...", Status: 410}
ErrTagNotFound   = &AppError{Code: "TAG_NOT_FOUND", Message: "...", Status: 404}
ErrValidation    = &AppError{Code: "VALIDATION_ERROR", Message: "...", Status: 400}
ErrRateLimited   = &AppError{Code: "RATE_LIMITED", Message: "...", Status: 429}
ErrInternal      = &AppError{Code: "INTERNAL_ERROR", Message: "...", Status: 500}
```

ハンドラー層で `errors.As` を使ってエラー型を判別し、統一フォーマットで JSON レスポンスを返す。

### pgtype 変換

Go の標準型と pgx の型の間の変換は `internal/service/pgconv.go` のヘルパー関数を使用:

| 関数 | 変換 |
|------|------|
| `ToPgTimestamptz(time.Time)` | `time.Time` → `pgtype.Timestamptz` |
| `ToPgDate(time.Time)` | `time.Time` → `pgtype.Date` |
| `ToPgText(string)` | `string` → `pgtype.Text` |
| `ToPgInt8(int64)` | `int64` → `pgtype.Int8` |
| `ToPgFloat8(*float64)` | `*float64` → `pgtype.Float8` |
| `FromPgTimestamptz(pgtype.Timestamptz)` | → `time.Time` |
| `PgTimestamptzPtr(pgtype.Timestamptz)` | → `*time.Time` (NULL → nil) |

---

## トラブルシューティング

### `web/dist/` が見つからないビルドエラー

```
pattern dist/*: no matching files found
```

**原因:** フロントエンドをビルドせずに Go バイナリをビルドした。

**解決:** 開発時は `-tags dev` を使用するか、先にフロントエンドをビルド:

```bash
# 方法1: 開発ビルドタグ
go build -tags dev -o memento ./cmd/memento

# 方法2: フロントエンドビルド後に本番ビルド
cd web && npm run build && cd ..
go build -o memento ./cmd/memento
```

### testcontainers-go のバージョンエラー

```
requires go >= 1.24
```

**原因:** testcontainers-go の最新版が Go 1.24 を要求。

**解決:** Go 1.23 互換の v0.35.0 を使用:

```bash
go get github.com/testcontainers/testcontainers-go@v0.35.0
go get github.com/testcontainers/testcontainers-go/modules/postgres@v0.35.0
```

### 統合テストで PostgreSQL 接続拒否

```
connection refused
```

**原因:** Docker-in-Docker 環境でのネットワーク疎通問題。

**解決:** 環境変数とホスト設定を追加:

```bash
docker run --rm \
  -e TESTCONTAINERS_RYUK_DISABLED=true \
  -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal \
  --add-host=host.docker.internal:host-gateway \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ...
```

### DB マイグレーションの失敗

**原因:** マイグレーションファイルの構文エラー、またはダウンマイグレーション未定義。

**解決:**
```bash
# マイグレーション状態を確認
goose -dir db/migrations postgres "$DATABASE_URL" status

# 問題のあるマイグレーションを1つ戻す
goose -dir db/migrations postgres "$DATABASE_URL" down

# 修正後に再実行
goose -dir db/migrations postgres "$DATABASE_URL" up
```

### WSL2 でのパフォーマンス問題

**原因:** Windows ファイルシステム (`/mnt/c/`) 上でのファイル I/O は遅い。

**解決:** ソースコードは必ず WSL のネイティブファイルシステム (`~/memento-memo`) に配置する。`/mnt/c/` は参照のみに使用。
