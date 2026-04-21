-- migrate:up

-- ledger.accounts: replace the integer flags bitfield with a TEXT[] constraints column.
-- Existing values:
--   0 -> '{}'                               (no constraints; e.g. escrow account)
--   1 -> '{"debits_must_not_exceed_credits"}' (user wallets)
--   2 -> '{"credits_must_not_exceed_debits"}' (points issuance account)
ALTER TABLE ledger.accounts ALTER COLUMN flags DROP DEFAULT;

ALTER TABLE ledger.accounts
    ALTER COLUMN flags TYPE TEXT[]
    USING CASE
        WHEN flags = 1 THEN ARRAY['debits_must_not_exceed_credits']
        WHEN flags = 2 THEN ARRAY['credits_must_not_exceed_debits']
        WHEN flags = 3 THEN ARRAY['debits_must_not_exceed_credits', 'credits_must_not_exceed_debits']
        ELSE '{}'
    END;

ALTER TABLE ledger.accounts ALTER COLUMN flags SET DEFAULT '{}';
ALTER TABLE ledger.accounts RENAME COLUMN flags TO constraints;

-- ledger.transfers: drop the flags column entirely.
-- The status column ('pending' | 'posted' | 'voided') already carries this information.
ALTER TABLE ledger.transfers DROP COLUMN flags;

-- migrate:down

ALTER TABLE ledger.transfers ADD COLUMN flags INT NOT NULL DEFAULT 0;

ALTER TABLE ledger.accounts RENAME COLUMN constraints TO flags;

ALTER TABLE ledger.accounts
    ALTER COLUMN flags TYPE INT
    USING CASE
        WHEN 'debits_must_not_exceed_credits' = ANY(flags) AND 'credits_must_not_exceed_debits' = ANY(flags) THEN 3
        WHEN 'debits_must_not_exceed_credits' = ANY(flags) THEN 1
        WHEN 'credits_must_not_exceed_debits' = ANY(flags) THEN 2
        ELSE 0
    END;

ALTER TABLE ledger.accounts ALTER COLUMN flags SET DEFAULT 0;
ALTER TABLE ledger.accounts ALTER COLUMN flags SET NOT NULL;
