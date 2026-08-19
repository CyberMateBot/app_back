CREATE TABLE IF NOT EXISTS user_feedback (
    id          BIGSERIAL PRIMARY KEY,
    profile_id  BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL CHECK (kind IN ('suggestion', 'bug')),
    message     TEXT NOT NULL CHECK (char_length(trim(message)) >= 3),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_feedback_kind_created_at
    ON user_feedback (kind, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_feedback_profile_id
    ON user_feedback (profile_id);
