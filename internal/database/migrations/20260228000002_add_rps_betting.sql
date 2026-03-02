-- migrate:up
-- Add betting columns to rps_games.
-- bet_amount: null means no bet on this game.
-- host_bet_transfer_id / guest_bet_transfer_id: pending ledger transfer IDs
-- created when each player places their bet escrow.
ALTER TABLE gaming.rps_games
    ADD COLUMN bet_amount            BIGINT,
    ADD COLUMN host_bet_transfer_id  UUID REFERENCES ledger.transfers(id),
    ADD COLUMN guest_bet_transfer_id UUID REFERENCES ledger.transfers(id);

CREATE INDEX idx_gaming_rps_games_bet_amount ON gaming.rps_games (bet_amount) WHERE bet_amount IS NOT NULL;
CREATE INDEX idx_gaming_rps_games_host_bet_transfer_id ON gaming.rps_games (host_bet_transfer_id) WHERE host_bet_transfer_id IS NOT NULL;
CREATE INDEX idx_gaming_rps_games_guest_bet_transfer_id ON gaming.rps_games (guest_bet_transfer_id) WHERE guest_bet_transfer_id IS NOT NULL;

-- migrate:down
DROP INDEX IF EXISTS gaming.idx_gaming_rps_games_guest_bet_transfer_id;
DROP INDEX IF EXISTS gaming.idx_gaming_rps_games_host_bet_transfer_id;
DROP INDEX IF EXISTS gaming.idx_gaming_rps_games_bet_amount;

ALTER TABLE gaming.rps_games
    DROP COLUMN IF EXISTS guest_bet_transfer_id,
    DROP COLUMN IF EXISTS host_bet_transfer_id,
    DROP COLUMN IF EXISTS bet_amount;
