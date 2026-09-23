-- Add token_version to admins table for JWT revocation support.
-- Incrementing this value invalidates all previously issued tokens for that admin.
ALTER TABLE admins ADD COLUMN IF NOT EXISTS token_version BIGINT NOT NULL DEFAULT 1;
