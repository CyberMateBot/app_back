-- Payment orders created for YooKassa checkout (coin packs and subscriptions).
-- One row per checkout attempt. idempotence_key is the value sent to YooKassa as
-- the Idempotence-Key header (also doubles as our own de-dup key on retries).
-- provider_payment_id is filled in once YooKassa responds to the create-payment
-- call and is the join key used by the webhook handler.
CREATE TABLE IF NOT EXISTS payments (
    id                    BIGSERIAL PRIMARY KEY,
    profile_id            BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    provider              TEXT NOT NULL DEFAULT 'yookassa',
    provider_payment_id   TEXT,
    idempotence_key       TEXT NOT NULL,
    kind                  TEXT NOT NULL CHECK (kind IN ('coin_pack', 'subscription')),
    item_id               TEXT NOT NULL,
    amount_rub            NUMERIC(12,2) NOT NULL,
    coins                 BIGINT NOT NULL DEFAULT 0,
    status                TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'succeeded', 'canceled', 'refunded')),
    confirmation_url      TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_idempotence_key ON payments (idempotence_key);
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_provider_payment_id ON payments (provider_payment_id) WHERE provider_payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payments_profile_id ON payments (profile_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status);
