package model

import "time"

// Task はPMツール横断の統一タスクモデル。
type Task struct {
	ID          string     `json:"id"`
	ExternalID  string     `json:"external_id"`
	Source      string     `json:"source"` // "clickup", "jira"
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	Assignee    string     `json:"assignee,omitempty"`
	StoryPoints *float64   `json:"story_points,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	Labels      []string   `json:"labels,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Project はPMツールのプロジェクト。
type Project struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
	Source     string `json:"source"`
	Name       string `json:"name"`
}

// Sprint はスプリントの統一モデル。
type Sprint struct {
	ID         string    `json:"id"`
	ExternalID string    `json:"external_id"`
	Source     string    `json:"source"`
	ProjectID  string    `json:"project_id"`
	Name       string    `json:"name"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Status     string    `json:"status"` // "active", "closed", "planned"
}

// TaskEvent はタスクのステータス変更イベント。
type TaskEvent struct {
	TaskID     string    `json:"task_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source"` // "api", "webhook", "time_in_status"
}

// Connection はPMツールへの接続情報。
type Connection struct {
	ID           string            `json:"id"`
	Source       string            `json:"source"`
	DisplayName  string            `json:"display_name"`
	APIKey       string            `json:"-"`
	Config       map[string]string `json:"config,omitempty"`
	SyncStatus   string            `json:"sync_status"`
	SyncError    *string           `json:"sync_error,omitempty"`
	LastSyncedAt *time.Time        `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// BurndownPoint はバーンダウンチャートの1データポイント。
type BurndownPoint struct {
	Date            time.Time `json:"date"`
	RemainingTasks  int       `json:"remaining_tasks"`
	RemainingPoints float64   `json:"remaining_points"`
	IdealRemaining  float64   `json:"ideal_remaining"`
}

// VelocityPoint はベロシティチャートの1スプリント分のデータ。
type VelocityPoint struct {
	SprintName      string  `json:"sprint_name"`
	CommittedPoints float64 `json:"committed_points"`
	CompletedPoints float64 `json:"completed_points"`
}

// CumulativeFlowPoint は累積フローダイアグラムの1日分のデータ。
type CumulativeFlowPoint struct {
	Date     time.Time      `json:"date"`
	Statuses map[string]int `json:"statuses"` // status名 -> タスク数
}

// LeadTimeStats はリードタイム/サイクルタイムの統計。
type LeadTimeStats struct {
	Percentile50 float64 `json:"p50_hours"`
	Percentile85 float64 `json:"p85_hours"`
	Percentile95 float64 `json:"p95_hours"`
	Average      float64 `json:"avg_hours"`
}
