-- migrate:up

-- Prevent the same ledger transfer from being assigned to two different games.
-- Without this, a bug that reuses a transfer ID would corrupt both games' audit trails.
ALTER TABLE gaming.rps_games
    ADD CONSTRAINT uq_rps_games_host_bet_transfer_id
        UNIQUE (host_bet_transfer_id),
    ADD CONSTRAINT uq_rps_games_guest_bet_transfer_id
        UNIQUE (guest_bet_transfer_id);

-- migrate:down

ALTER TABLE gaming.rps_games
    DROP CONSTRAINT IF EXISTS uq_rps_games_host_bet_transfer_id,
    DROP CONSTRAINT IF EXISTS uq_rps_games_guest_bet_transfer_id;
