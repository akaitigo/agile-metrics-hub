package adapter

import (
	"context"

	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

// PMToolAdapter は各プロジェクト管理ツールのAPI差分を吸収する共通インターフェース。
// 新規PMツール追加時はこのインターフェースを実装する。
type PMToolAdapter interface {
	// Name はアダプターの識別名を返す（例: "clickup", "jira"）
	Name() string

	// FetchTasks は指定プロジェクトの全タスクを取得する
	FetchTasks(ctx context.Context, projectID string) ([]model.Task, error)

	// FetchSprints は指定プロジェクトのスプリント一覧を取得する
	FetchSprints(ctx context.Context, projectID string) ([]model.Sprint, error)

	// FetchTaskHistory はタスクのステータス変更履歴を取得する
	// PMツールが履歴APIを提供しない場合は ErrHistoryNotSupported を返す
	FetchTaskHistory(ctx context.Context, taskID string) ([]model.TaskEvent, error)
}
