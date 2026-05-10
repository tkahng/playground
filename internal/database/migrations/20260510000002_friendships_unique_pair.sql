-- migrate:up
CREATE UNIQUE INDEX gaming_friendships_unique_pair
ON gaming.friendships (
    LEAST(requesting_player_id::text, invited_player_id::text),
    GREATEST(requesting_player_id::text, invited_player_id::text)
);

-- migrate:down
DROP INDEX gaming.gaming_friendships_unique_pair;
