-- 001_initial_schema.sql
-- Agile Metrics Hub: コアデータモデル

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- PMツール接続情報（APIキーは暗号化保存）
CREATE TABLE connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source VARCHAR(32) NOT NULL CHECK (source IN ('clickup', 'jira', 'linear', 'asana')),
    display_name VARCHAR(256) NOT NULL,
    encrypted_api_key BYTEA NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    last_synced_at TIMESTAMPTZ,
    sync_status VARCHAR(32) NOT NULL DEFAULT 'pending' CHECK (sync_status IN ('pending', 'syncing', 'success', 'error')),
    sync_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- PMツールのプロジェクト（ClickUp: List, Jira: Board/Project）
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    external_id VARCHAR(256) NOT NULL,
    source VARCHAR(32) NOT NULL,
    name VARCHAR(512) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (connection_id, external_id)
);

CREATE INDEX idx_projects_connection ON projects(connection_id);

-- 統一タスクモデル
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    external_id VARCHAR(256) NOT NULL,
    source VARCHAR(32) NOT NULL,
    title VARCHAR(1024) NOT NULL,
    status VARCHAR(256) NOT NULL,
    assignee VARCHAR(256),
    story_points DOUBLE PRECISION,
    priority VARCHAR(64),
    labels TEXT[],
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    raw_data JSONB,
    UNIQUE (project_id, external_id)
);

CREATE INDEX idx_tasks_project ON tasks(project_id);
CREATE INDEX idx_tasks_status ON tasks(project_id, status);
CREATE INDEX idx_tasks_updated ON tasks(project_id, updated_at);

-- スプリント情報
CREATE TABLE sprints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    external_id VARCHAR(256) NOT NULL,
    source VARCHAR(32) NOT NULL,
    name VARCHAR(512) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    status VARCHAR(32) NOT NULL CHECK (status IN ('planned', 'active', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, external_id)
);

CREATE INDEX idx_sprints_project ON sprints(project_id);
CREATE INDEX idx_sprints_status ON sprints(project_id, status);

-- スプリントとタスクの関連（多対多）
CREATE TABLE sprint_tasks (
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    PRIMARY KEY (sprint_id, task_id)
);

-- タスクのステータス変更履歴
CREATE TABLE task_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    from_status VARCHAR(256),
    to_status VARCHAR(256) NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL,
    source VARCHAR(32) NOT NULL CHECK (source IN ('api', 'webhook', 'time_in_status')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_events_task ON task_events(task_id, changed_at);
CREATE INDEX idx_task_events_timestamp ON task_events(changed_at);

-- 日次タスク状態スナップショット（バーンダウン/累積フロー用）
CREATE TABLE snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    snapshot_date DATE NOT NULL,
    total_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    total_points DOUBLE PRECISION NOT NULL DEFAULT 0,
    completed_points DOUBLE PRECISION NOT NULL DEFAULT 0,
    status_counts JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, snapshot_date)
);

CREATE INDEX idx_snapshots_project_date ON snapshots(project_id, snapshot_date);

-- 更新日時の自動更新トリガー
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_connections_updated_at
    BEFORE UPDATE ON connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
