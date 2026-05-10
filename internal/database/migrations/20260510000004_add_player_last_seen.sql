-- migrate:up
ALTER TABLE gaming.players
    ADD COLUMN last_seen_at TIMESTAMPTZ;

CREATE INDEX idx_players_last_seen_at ON gaming.players (last_seen_at);

-- migrate:down
DROP INDEX IF EXISTS gaming.idx_players_last_seen_at;
ALTER TABLE gaming.players
    DROP COLUMN last_seen_at;
