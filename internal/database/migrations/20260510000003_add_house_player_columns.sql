-- migrate:up
ALTER TABLE gaming.players
    ADD COLUMN is_house          BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN last_house_game_at TIMESTAMPTZ;

-- migrate:down
ALTER TABLE gaming.players
    DROP COLUMN is_house,
    DROP COLUMN last_house_game_at;
