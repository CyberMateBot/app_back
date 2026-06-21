-- Allow system generation debits without an admin actor.
ALTER TABLE token_transactions
    ALTER COLUMN admin_id DROP NOT NULL;
