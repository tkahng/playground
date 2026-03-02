-- migrate:up
-- ledger schema
CREATE SCHEMA IF NOT EXISTS ledger;

-- ledger.accounts
-- Tracks all accounts in the double-entry ledger.
-- Every monetary value in the system lives in an account.
-- Balance = credits_posted - debits_posted (for credit-normal accounts like user wallets)
-- Available balance = (credits_posted - debits_posted) - debits_pending
CREATE TABLE ledger.accounts (
    id              UUID        PRIMARY KEY DEFAULT uuidv7(),
    code            TEXT        NOT NULL UNIQUE,
    entity_type     TEXT        NOT NULL,
    entity_id       UUID,
    ledger_code     TEXT        NOT NULL DEFAULT 'POINTS',
    -- flags bitfield:
    --   DEBITS_MUST_NOT_EXCEED_CREDITS = 1  (prevents overdraft; use on user wallets)
    --   CREDITS_MUST_NOT_EXCEED_DEBITS = 2  (use on issuance accounts)
    flags           INT         NOT NULL DEFAULT 0,
    debits_pending  BIGINT      NOT NULL DEFAULT 0 CHECK (debits_pending >= 0),
    credits_pending BIGINT      NOT NULL DEFAULT 0 CHECK (credits_pending >= 0),
    debits_posted   BIGINT      NOT NULL DEFAULT 0 CHECK (debits_posted >= 0),
    credits_posted  BIGINT      NOT NULL DEFAULT 0 CHECK (credits_posted >= 0),
    metadata        JSONB       NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_ledger_accounts_code ON ledger.accounts (code);
CREATE INDEX idx_ledger_accounts_entity ON ledger.accounts (entity_type, entity_id);
CREATE INDEX idx_ledger_accounts_ledger_code ON ledger.accounts (ledger_code);

CREATE TRIGGER handle_ledger_accounts_updated_at
BEFORE UPDATE ON ledger.accounts
FOR EACH ROW EXECUTE PROCEDURE utility.set_current_timestamp_updated_at();

-- ledger.transfers
-- Every value movement is a transfer: debit one account, credit another.
-- Conservation: sum of all debits = sum of all credits (per ledger_code).
-- status: 'pending' | 'posted' | 'voided'
-- flags bitfield:
--   PENDING          = 1  (two-phase pending hold)
--   POST_PENDING     = 2  (posts a pending transfer)
--   VOID_PENDING     = 4  (voids a pending transfer)
--   LINKED           = 8  (part of an atomic linked group)
-- transfer_code examples: 'purchase' | 'bet_escrow' | 'bet_win' | 'bet_refund' | 'bet_void'
-- reference_type examples: 'rps_game' | 'stripe_checkout'
CREATE TABLE ledger.transfers (
    id                UUID        PRIMARY KEY DEFAULT uuidv7(),
    ledger_code       TEXT        NOT NULL DEFAULT 'POINTS',
    debit_account_id  UUID        NOT NULL REFERENCES ledger.accounts(id),
    credit_account_id UUID        NOT NULL REFERENCES ledger.accounts(id),
    amount            BIGINT      NOT NULL CHECK (amount > 0),
    pending_id        UUID        REFERENCES ledger.transfers(id),
    flags             INT         NOT NULL DEFAULT 0,
    status            TEXT        NOT NULL DEFAULT 'posted' CHECK (status IN ('pending', 'posted', 'voided')),
    transfer_code     TEXT        NOT NULL,
    reference_type    TEXT,
    reference_id      UUID,
    timeout_at        TIMESTAMPTZ,
    metadata          JSONB       NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_ledger_transfers_debit_account ON ledger.transfers (debit_account_id);
CREATE INDEX idx_ledger_transfers_credit_account ON ledger.transfers (credit_account_id);
CREATE INDEX idx_ledger_transfers_reference ON ledger.transfers (reference_type, reference_id);
CREATE INDEX idx_ledger_transfers_pending_id ON ledger.transfers (pending_id) WHERE pending_id IS NOT NULL;
CREATE INDEX idx_ledger_transfers_status ON ledger.transfers (status);
CREATE INDEX idx_ledger_transfers_ledger_code ON ledger.transfers (ledger_code);

CREATE TRIGGER handle_ledger_transfers_updated_at
BEFORE UPDATE ON ledger.transfers
FOR EACH ROW EXECUTE PROCEDURE utility.set_current_timestamp_updated_at();

-- Seed system accounts required by the ledger service
INSERT INTO ledger.accounts (code, entity_type, ledger_code, flags)
VALUES
    ('system:points_issuance', 'system', 'POINTS', 2),
    ('system:game_escrow',     'system', 'POINTS', 0);

-- migrate:down
DROP TRIGGER IF EXISTS handle_ledger_transfers_updated_at ON ledger.transfers;
DROP INDEX IF EXISTS ledger.idx_ledger_transfers_ledger_code;
DROP INDEX IF EXISTS ledger.idx_ledger_transfers_status;
DROP INDEX IF EXISTS ledger.idx_ledger_transfers_pending_id;
DROP INDEX IF EXISTS ledger.idx_ledger_transfers_reference;
DROP INDEX IF EXISTS ledger.idx_ledger_transfers_credit_account;
DROP INDEX IF EXISTS ledger.idx_ledger_transfers_debit_account;
DROP TABLE IF EXISTS ledger.transfers;

DROP TRIGGER IF EXISTS handle_ledger_accounts_updated_at ON ledger.accounts;
DROP INDEX IF EXISTS ledger.idx_ledger_accounts_ledger_code;
DROP INDEX IF EXISTS ledger.idx_ledger_accounts_entity;
DROP INDEX IF EXISTS ledger.idx_ledger_accounts_code;
DROP TABLE IF EXISTS ledger.accounts;

DROP SCHEMA IF EXISTS ledger;
