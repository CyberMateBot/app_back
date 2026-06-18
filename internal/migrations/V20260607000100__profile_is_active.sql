-- Admin panel: block/unblock Mini App users.
ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_profiles_is_active ON profiles (is_active);
CREATE INDEX IF NOT EXISTS idx_profiles_created_at ON profiles (created_at DESC);
