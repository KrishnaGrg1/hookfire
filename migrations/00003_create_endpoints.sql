-- +goose Up
CREATE TABLE endpoints (
  id         BIGSERIAL PRIMARY KEY,
  app_id     BIGINT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
  url        TEXT NOT NULL,
  secret     TEXT NOT NULL,
  is_active  BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS endpoints;
