
-- name: CreateEndpoint :one
INSERT INTO endpoints (app_id, url, secret)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListEndpointsByApp :many
SELECT * FROM endpoints
WHERE app_id = $1 AND is_active = true;

-- name: GetEndpointByID :one
SELECT * FROM endpoints
WHERE id = $1;

-- name: DeleteEndpoint :exec
UPDATE endpoints SET is_active = false
WHERE id = $1 AND app_id = $2;
