-- name: UpsertSnapshot :one
INSERT INTO snapshots (project_id, snapshot_date, total_tasks, completed_tasks, total_points, completed_points, status_counts)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (project_id, snapshot_date)
DO UPDATE SET
    total_tasks = EXCLUDED.total_tasks,
    completed_tasks = EXCLUDED.completed_tasks,
    total_points = EXCLUDED.total_points,
    completed_points = EXCLUDED.completed_points,
    status_counts = EXCLUDED.status_counts
RETURNING *;

-- name: ListSnapshotsByDateRange :many
SELECT * FROM snapshots
WHERE project_id = $1
  AND snapshot_date >= $2
  AND snapshot_date <= $3
ORDER BY snapshot_date;
