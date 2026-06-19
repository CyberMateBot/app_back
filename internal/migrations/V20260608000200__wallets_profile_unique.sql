-- One wallet per profile (prevents balance split across duplicate rows).
DELETE FROM wallets w1
USING wallets w2
WHERE w1.profile_id = w2.profile_id
  AND w1.id > w2.id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_profile_id_unique ON wallets (profile_id);
