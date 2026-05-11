-- migrate:up
-- Enforce at the database level that each player may be the guest
-- (invited/pending participant) in at most one pending game at a time.
--
-- The host's participant always has status='completed' on game creation,
-- so the host side is enforced by the application-level check inside the
-- RequestGame transaction. This index covers the guest side and also acts
-- as a safety net if the application check is bypassed.
CREATE UNIQUE INDEX idx_rps_participants_one_pending_per_player
    ON gaming.rps_participants (player_id)
    WHERE status = 'pending';

-- migrate:down
DROP INDEX IF EXISTS gaming.idx_rps_participants_one_pending_per_player;
