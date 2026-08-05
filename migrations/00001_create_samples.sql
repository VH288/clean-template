-- +goose Up
CREATE TABLE IF NOT EXISTS samples (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_samples_status ON samples(status);
CREATE INDEX IF NOT EXISTS idx_samples_created_at ON samples(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_samples_created_at;
DROP INDEX IF EXISTS idx_samples_status;
DROP TABLE IF EXISTS samples;
