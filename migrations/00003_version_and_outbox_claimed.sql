-- +goose Up
ALTER TABLE samples ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_outbox_processing ON outbox_events(status, claimed_at)
    WHERE status = 'processing';

-- +goose Down
DROP INDEX IF EXISTS idx_outbox_processing;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS claimed_at;
ALTER TABLE samples DROP COLUMN IF EXISTS version;
