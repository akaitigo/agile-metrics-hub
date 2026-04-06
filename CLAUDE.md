# Agile Metrics Hub

PMツール横断のアジャイルメトリクスダッシュボード

## 技術スタック
- Frontend: Next.js (TypeScript) + shadcn/ui + Recharts
- Backend: Go + sqlc
- DB: PostgreSQL
- Infra: GCP Cloud Run

## コマンド
```
make check       # lint → test → build (frontend + backend)
make quality     # 品質ゲート
make fe-dev      # frontend dev server
make be-dev      # backend dev server
```

## ルール
- TypeScript: ~/.claude/rules/typescript.md 参照
- Go: ~/.claude/rules/go.md 参照
- セキュリティ: ~/.claude/rules/security.md 参照

## 構造
```
frontend/   Next.js App Router (src/app/)
backend/    Go API (cmd/server/, internal/)
docs/       ADR, 品質チェックリスト
```

## 禁止事項
- any型(TS) / console.log / TODO コミット
- .env・credentials のコミット
- lint設定の無効化（ADR必須）

## 状態管理
- git log + GitHub Issues でセッション間の状態を管理
- セッション開始: `bash .claude/startup.sh`
