-- migrate:up
-- Enforce that at most one pending rematch request exists per game.
-- The application-level duplicate check in RequestRematch runs inside a
-- transaction, but this index is the database-level backstop that prevents
-- two concurrent callers from both slipping past that check.
CREATE UNIQUE INDEX idx_rps_rematch_requests_one_pending_per_game
    ON gaming.rps_rematch_requests (original_game_id)
    WHERE status = 'pending';

-- migrate:down
DROP INDEX IF EXISTS gaming.idx_rps_rematch_requests_one_pending_per_game;
