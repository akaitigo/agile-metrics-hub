# Agile Metrics Hub — アーキテクチャ概要

## Adapter パターン

各PMツール（ClickUp, Jira, Linear, Asana）のAPI差分を吸収する共通インターフェース。
新規PMツール追加時は `backend/internal/adapter/` に新しいアダプターを実装するのみ。

```
PMToolAdapter interface {
    FetchTasks(ctx, projectID) -> []Task
    FetchSprints(ctx, projectID) -> []Sprint
    SyncWebhook(ctx, payload) -> []TaskEvent
}
```

## データフロー

1. ユーザーがAPIキーを設定 → `connections` テーブルに暗号化保存
2. 初回同期: Adapter経由で全タスクを取得 → `tasks` テーブルに保存
3. 定期同期: Webhook + ポーリングのハイブリッドで差分更新
4. スナップショット: 日次でタスク状態のスナップショットを `snapshots` テーブルに保存
5. メトリクス計算: `metrics` パッケージがスナップショットからチャートデータを生成

## 外部サービス連携

| サービス | 用途 | 認証方式 |
|---------|------|---------|
| ClickUp API v2 | タスク/スプリントデータ取得 | Personal API Token |
| Jira Cloud REST API | タスク/スプリントデータ取得 | API Token (email + token) |
| PostgreSQL | データ永続化 | 接続文字列 (環境変数) |

## 設計判断

- docs/adr/ を参照
