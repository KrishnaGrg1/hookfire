-- name: CreateAttempt :one
INSERT INTO delivery_attempts (event_id, endpoint_id, status, http_status, attempt_number, next_retry_at, delivered_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAttemptsByEvent :many
SELECT * FROM delivery_attempts
WHERE event_id = $1
ORDER BY created_at DESC;
