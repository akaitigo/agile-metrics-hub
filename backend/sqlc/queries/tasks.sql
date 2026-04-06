-- name: UpsertTask :one
INSERT INTO tasks (project_id, external_id, source, title, status, assignee, story_points, priority, labels, created_at, updated_at, completed_at, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (project_id, external_id)
DO UPDATE SET
    title = EXCLUDED.title,
    status = EXCLUDED.status,
    assignee = EXCLUDED.assignee,
    story_points = EXCLUDED.story_points,
    priority = EXCLUDED.priority,
    labels = EXCLUDED.labels,
    updated_at = EXCLUDED.updated_at,
    completed_at = EXCLUDED.completed_at,
    raw_data = EXCLUDED.raw_data
RETURNING *;

-- name: ListTasksByProject :many
SELECT * FROM tasks WHERE project_id = $1 ORDER BY updated_at DESC;

-- name: CountTasksByStatus :many
SELECT status, COUNT(*) as count
FROM tasks
WHERE project_id = $1
GROUP BY status;

-- name: GetTaskByExternalID :one
SELECT * FROM tasks WHERE project_id = $1 AND external_id = $2;
