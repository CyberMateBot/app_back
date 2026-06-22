-- Per-user subscription state. A row exists only for users who were granted a
-- paid plan (or had one granted by admin). Absence of a row means the "free" plan.
CREATE TABLE IF NOT EXISTS user_subscriptions (
    profile_id  BIGINT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    plan_id     TEXT NOT NULL DEFAULT 'free',
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ,
    granted_by  BIGINT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_expires_at
    ON user_subscriptions (expires_at);
