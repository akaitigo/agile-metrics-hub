-- name: CreateConnection :one
INSERT INTO connections (source, display_name, encrypted_api_key, config)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetConnection :one
SELECT * FROM connections WHERE id = $1;

-- name: ListConnections :many
SELECT * FROM connections ORDER BY created_at DESC;

-- name: UpdateConnectionSyncStatus :exec
UPDATE connections
SET sync_status = $2, sync_error = $3, last_synced_at = CASE WHEN $2 = 'success' THEN NOW() ELSE last_synced_at END
WHERE id = $1;

-- name: DeleteConnection :exec
DELETE FROM connections WHERE id = $1;
