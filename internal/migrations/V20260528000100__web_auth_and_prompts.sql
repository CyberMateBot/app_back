-- Web site auth + prompts storage (separate from Telegram profiles).
-- Passwords are stored as bcrypt hashes.

CREATE TABLE IF NOT EXISTS web_accounts (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_web_accounts_created_at ON web_accounts(created_at);

CREATE TABLE IF NOT EXISTS web_prompt_history (
    id BIGSERIAL PRIMARY KEY,
    web_account_id BIGINT NOT NULL REFERENCES web_accounts(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_web_prompt_history_account_created_at
    ON web_prompt_history(web_account_id, created_at DESC);

