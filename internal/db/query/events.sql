-- name: CreateEvent :one
INSERT INTO events (app_id, event_type, payload)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEventByID :one
SELECT * FROM events
WHERE id = $1;

-- name: ListEventsByApp :many
SELECT * FROM events
WHERE app_id = $1
ORDER BY created_at DESC
LIMIT 50;
