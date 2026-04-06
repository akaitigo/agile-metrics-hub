# Changelog

## v1.0.0 (2026-04-06)

MVP リリース — PMツール横断のアジャイルメトリクスダッシュボード

### Features
- ClickUp API v2 アダプター（タスク取得、Time in Status履歴）
- Jira Cloud REST API v3 アダプター（Issue取得、スプリント管理、changelog履歴）
- バーンダウンチャート（実績線 + 理想線）
- ベロシティチャート（コミット vs 完了 + 平均算出）
- 累積フローダイアグラム（ステータス別日次推移）
- リードタイム分析（P50 / P85 / P95 パーセンタイル）
- ダッシュボードUI（2x2グリッドレイアウト、Recharts）
- 接続テストエンドポイント（POST /api/connections/test）
- APIキー暗号化（AES-256-GCM）
- CORS ミドルウェア（環境変数制御）

### Infrastructure
- GitHub Actions CI（Frontend + Backend 並列ジョブ）
- Docker Compose 開発環境（PostgreSQL 16 + pgAdmin）
- Dependabot + auto-merge
- lefthook pre-commit hooks（lint / format / test）
- biome + oxlint（Frontend）、golangci-lint（Backend）

### Architecture Decisions
- ADR 001: コアデータモデル設計（Adapter パターン + sqlc）
- ADR 002: 認証方式選定（MVP: APIキー直接入力）
- ADR 003: データ同期戦略（Webhook + ポーリング ハイブリッド）
- ADR 004: MVP段階でのデモデータ使用
