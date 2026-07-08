-- name: CreateTaskEvent :one
INSERT INTO task_events (task_id, from_status, to_status, changed_at, source)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListTaskEvents :many
SELECT * FROM task_events WHERE task_id = $1 ORDER BY changed_at;

-- name: ListProjectTaskEvents :many
SELECT te.* FROM task_events te
JOIN tasks t ON t.id = te.task_id
WHERE t.project_id = $1
  AND te.changed_at >= $2
  AND te.changed_at <= $3
ORDER BY te.changed_at;
