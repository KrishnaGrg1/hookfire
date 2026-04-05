-- name: CreateApplication :one
INSERT INTO applications(name,api_key)
VALUES($1,$2)
RETURNING id, name, api_key, created_at;