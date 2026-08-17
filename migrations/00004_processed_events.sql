-- +goose Up
CREATE TABLE IF NOT EXISTS processed_events (
    event_id UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_processed_events_processed_at ON processed_events(processed_at);

-- +goose Down
DROP INDEX IF EXISTS idx_processed_events_processed_at;
DROP TABLE IF EXISTS processed_events;
