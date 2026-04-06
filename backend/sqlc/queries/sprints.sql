-- name: UpsertSprint :one
INSERT INTO sprints (project_id, external_id, source, name, start_date, end_date, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (project_id, external_id)
DO UPDATE SET name = EXCLUDED.name, start_date = EXCLUDED.start_date, end_date = EXCLUDED.end_date, status = EXCLUDED.status
RETURNING *;

-- name: ListSprintsByProject :many
SELECT * FROM sprints WHERE project_id = $1 ORDER BY start_date DESC;

-- name: GetActiveSprint :one
SELECT * FROM sprints WHERE project_id = $1 AND status = 'active' LIMIT 1;

-- name: ListSprintTasks :many
SELECT t.* FROM tasks t
JOIN sprint_tasks st ON st.task_id = t.id
WHERE st.sprint_id = $1
ORDER BY t.updated_at DESC;

-- name: AddSprintTask :exec
INSERT INTO sprint_tasks (sprint_id, task_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
