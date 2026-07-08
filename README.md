# Agile Metrics Hub

> PMツール横断のアジャイルメトリクスダッシュボード

ClickUp / Jira 等のAPIキーを設定するだけで、バーンダウン・ベロシティ・累積フロー・リードタイム分析を統一ダッシュボードで表示。

## 主な機能

- **バーンダウンチャート** — スプリントの残ポイント推移（実績 vs 理想線）
- **ベロシティチャート** — スプリント別のコミット vs 完了ポイント
- **累積フローダイアグラム** — ステータス別タスク数の日次推移
- **リードタイム分析** — パーセンタイル表示（P50 / P85 / P95）
- **PMツール横断比較** — 複数ツールのメトリクスを並列表示

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| Frontend | Next.js (TypeScript) + Recharts |
| Backend | Go |
| Database | PostgreSQL 16 |
| Infra | GCP Cloud Run |

## クイックスタート

### 前提条件

- Node.js >= 22
- Go >= 1.23
- Docker & Docker Compose（PostgreSQL用）
- pnpm

### 1. クローン & セットアップ

```bash
git clone git@github.com:akaitigo/agile-metrics-hub.git
cd agile-metrics-hub
cp .env.example .env
```

### 2. 依存インストール

```bash
# Frontend
cd frontend && pnpm install && cd ..

# Backend
cd backend && go mod tidy && cd ..
```

### 3. DB起動

```bash
docker compose up -d
```

### 4. 開発サーバー起動

```bash
# ターミナル1: Backend API (http://localhost:8080)
make be-dev

# ターミナル2: Frontend (http://localhost:3000)
make fe-dev
```

### 5. ダッシュボード表示

ブラウザで http://localhost:3000/dashboard を開く。デモデータで4種のチャートが表示されます。

## API エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/health` | ヘルスチェック |
| POST | `/api/connections/test` | PMツール接続テスト |
| GET | `/api/metrics/burndown?project_id=X` | バーンダウンデータ |
| GET | `/api/metrics/velocity?count=5` | ベロシティデータ |
| GET | `/api/metrics/cumulative-flow?days=30` | 累積フローデータ |
| GET | `/api/metrics/lead-time?project_id=X` | リードタイム統計 |

## 開発

```bash
make check    # format → lint → typecheck → test → build
make quality  # 品質ゲート
```

## 設計判断

アーキテクチャ判断は `docs/adr/` に記録:

- [ADR 001](docs/adr/001-data-model.md) — コアデータモデル設計
- [ADR 002](docs/adr/002-auth-strategy.md) — 認証方式の選定
- [ADR 003](docs/adr/003-sync-strategy.md) — データ同期戦略
- [ADR 004](docs/adr/004-mvp-demo-data.md) — MVP段階でのデモデータ使用

## ライセンス

MIT
