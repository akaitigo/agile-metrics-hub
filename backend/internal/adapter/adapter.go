package adapter

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

var (
	ErrHistoryNotSupported = errors.New("task history API not supported by this adapter")
	ErrSprintsNotSupported = errors.New("sprints not supported by this adapter")
	ErrRateLimited         = errors.New("rate limit exceeded")
	ErrUnauthorized        = errors.New("invalid API credentials")
	ErrNotFound            = errors.New("resource not found")
)

// PMToolAdapter は各プロジェクト管理ツールのAPI差分を吸収する共通インターフェース。
type PMToolAdapter interface {
	// Name はアダプターの識別名を返す（例: "clickup", "jira"）
	Name() string

	// TestConnection は接続の有効性を確認する
	TestConnection(ctx context.Context) error

	// FetchProjects は利用可能なプロジェクト一覧を取得する
	FetchProjects(ctx context.Context) ([]model.Project, error)

	// FetchTasks は指定プロジェクトの全タスクを取得する
	FetchTasks(ctx context.Context, projectID string) ([]model.Task, error)

	// FetchSprints は指定プロジェクトのスプリント一覧を取得する
	FetchSprints(ctx context.Context, projectID string) ([]model.Sprint, error)

	// FetchTaskHistory はタスクのステータス変更履歴を取得する
	// PMツールが履歴APIを提供しない場合は ErrHistoryNotSupported を返す
	FetchTaskHistory(ctx context.Context, taskID string) ([]model.TaskEvent, error)
}

// AdapterFactory はAPIキーと設定からアダプターインスタンスを生成する関数型。
type AdapterFactory func(apiKey string, config map[string]string) (PMToolAdapter, error)

// Registry はアダプターファクトリーを管理するレジストリ。
type Registry struct {
	mu        sync.RWMutex
	factories map[string]AdapterFactory
}

// NewRegistry は新しいレジストリを作成する。
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]AdapterFactory),
	}
}

// Register はアダプターファクトリーをレジストリに登録する。
func (r *Registry) Register(source string, factory AdapterFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[source] = factory
}

// Create は指定されたソースのアダプターインスタンスを生成する。
func (r *Registry) Create(source, apiKey string, config map[string]string) (PMToolAdapter, error) {
	r.mu.RLock()
	factory, ok := r.factories[source]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unsupported PM tool: %s", source)
	}
	return factory(apiKey, config)
}

// Sources は登録済みのソース一覧を返す。
func (r *Registry) Sources() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sources := make([]string, 0, len(r.factories))
	for s := range r.factories {
		sources = append(sources, s)
	}
	return sources
}
