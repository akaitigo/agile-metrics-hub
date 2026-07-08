-- name: UpsertProject :one
INSERT INTO projects (connection_id, external_id, source, name, config)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (connection_id, external_id)
DO UPDATE SET name = EXCLUDED.name, config = EXCLUDED.config
RETURNING *;

-- name: ListProjectsByConnection :many
SELECT * FROM projects WHERE connection_id = $1 ORDER BY name;

-- name: ListAllProjects :many
SELECT * FROM projects ORDER BY name;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1;
