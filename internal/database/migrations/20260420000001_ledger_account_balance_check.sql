-- migrate:up

-- Enforce overdraft protection at the DB level for accounts that carry the
-- debits_must_not_exceed_credits constraint (i.e. user wallets).
-- Mirrors checkAvailableBalanceConstraint in the application layer:
--   available_balance = credits_posted - debits_posted - debits_pending >= 0
-- This fires on every UPDATE, catching any path that bypasses app-layer checks.
ALTER TABLE ledger.accounts
    ADD CONSTRAINT chk_accounts_available_balance CHECK (
        NOT ('debits_must_not_exceed_credits' = ANY(constraints))
        OR (credits_posted - debits_posted - debits_pending) >= 0
    );

-- migrate:down

ALTER TABLE ledger.accounts
    DROP CONSTRAINT IF EXISTS chk_accounts_available_balance;
