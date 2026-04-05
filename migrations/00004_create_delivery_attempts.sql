-- +goose Up
CREATE TABLE delivery_attempts (
  id             BIGSERIAL PRIMARY KEY,
  event_id       BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  endpoint_id    BIGINT NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
  status         TEXT NOT NULL DEFAULT 'pending',
  http_status    INT,
  attempt_number INT NOT NULL DEFAULT 1,
  next_retry_at  TIMESTAMPTZ,
  delivered_at   TIMESTAMPTZ,
  created_at     TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_attempts_event_id ON delivery_attempts(event_id);
CREATE INDEX idx_attempts_status ON delivery_attempts(status);

-- +goose Down
DROP TABLE IF EXISTS delivery_attempts;
