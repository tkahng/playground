-- migrate:up
CREATE TABLE gaming.rps_rematch_requests (
    id                   UUID        NOT NULL DEFAULT uuidv7() PRIMARY KEY,
    original_game_id     UUID        NOT NULL REFERENCES gaming.rps_games(id),
    requesting_player_id UUID        NOT NULL REFERENCES gaming.players(id),
    invited_player_id    UUID        NOT NULL REFERENCES gaming.players(id),
    status               TEXT        NOT NULL DEFAULT 'pending'
                             CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),
    new_game_id          UUID        REFERENCES gaming.rps_games(id),
    expires_at           TIMESTAMPTZ NOT NULL,
    metadata             JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_rps_rematch_requests_original_game_id ON gaming.rps_rematch_requests (original_game_id);
CREATE INDEX idx_rps_rematch_requests_status           ON gaming.rps_rematch_requests (status);
CREATE INDEX idx_rps_rematch_requests_expires_at       ON gaming.rps_rematch_requests (expires_at);

-- migrate:down
DROP TABLE IF EXISTS gaming.rps_rematch_requests;
