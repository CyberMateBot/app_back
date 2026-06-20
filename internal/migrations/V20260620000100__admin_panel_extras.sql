-- Admin panel: broadcast history, app settings, model pricing overrides.

CREATE TABLE IF NOT EXISTS admin_broadcasts (
    id            BIGSERIAL PRIMARY KEY,
    admin_id      BIGINT NOT NULL REFERENCES admins(id),
    message       TEXT NOT NULL,
    target        VARCHAR(20) NOT NULL,
    parse_mode    VARCHAR(20) NOT NULL DEFAULT 'HTML',
    sent_count    INTEGER NOT NULL DEFAULT 0,
    failed_count  INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_broadcasts_created_at
    ON admin_broadcasts (created_at DESC);

CREATE TABLE IF NOT EXISTS admin_settings (
    key         TEXT PRIMARY KEY,
    value       JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS model_configs (
    model_id     TEXT PRIMARY KEY,
    category     TEXT NOT NULL,
    name         TEXT NOT NULL,
    provider     TEXT NOT NULL DEFAULT '',
    price_coins  INTEGER NOT NULL DEFAULT 0 CHECK (price_coins >= 0),
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
