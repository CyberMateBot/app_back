-- Admin token operations log (profile_id = admin API user id).
CREATE TABLE IF NOT EXISTS token_transactions (
    id            BIGSERIAL PRIMARY KEY,
    profile_id    BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    admin_id      BIGINT NOT NULL REFERENCES admins(id),
    operation     VARCHAR(10) NOT NULL CHECK (operation IN ('credit', 'debit')),
    amount        INTEGER NOT NULL CHECK (amount > 0),
    balance_after INTEGER NOT NULL CHECK (balance_after >= 0),
    reason        VARCHAR(255),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_token_transactions_profile_created_at
    ON token_transactions (profile_id, created_at DESC);
