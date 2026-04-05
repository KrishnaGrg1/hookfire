-- +goose Up
CREATE TABLE IF NOT EXISTS applications(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    api_key TEXT UNIQUE NOT NULL,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS applications;

