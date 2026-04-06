# Agile Metrics Hub

PMツール横断のアジャイルメトリクスダッシュボード。ClickUp / Jira 等のAPIキーを設定するだけで、バーンダウン・ベロシティ・累積フロー・リードタイム分析を統一表示。

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| Frontend | Next.js (TypeScript) + Recharts |
| Backend | Go |
| Database | PostgreSQL |
| Infra | GCP Cloud Run |

## セットアップ

### 前提条件

- Node.js >= 22
- Go >= 1.23
- PostgreSQL
- pnpm

### インストール

```bash
# リポジトリをクローン
git clone git@github.com:akaitigo/agile-metrics-hub.git
cd agile-metrics-hub

# 環境変数を設定
cp .env.example .env

# Frontend
cd frontend && pnpm install && cd ..

# Backend
cd backend && go mod tidy && cd ..
```

### 開発サーバー起動

```bash
# Frontend (http://localhost:3000)
make fe-dev

# Backend (http://localhost:8080)
make be-dev
```

### テスト・ビルド

```bash
make check    # format → lint → typecheck → test → build
make quality  # 品質ゲート
```

## ライセンス

MIT
