-- migrate:up
ALTER TABLE gaming.friendships DROP CONSTRAINT friendships_status_check;
ALTER TABLE gaming.friendships ADD CONSTRAINT friendships_status_check CHECK (
    status IN ('pending', 'accepted', 'declined', 'blocked')
);

-- migrate:down
UPDATE gaming.friendships SET status = 'declined' WHERE status = 'blocked';
ALTER TABLE gaming.friendships DROP CONSTRAINT friendships_status_check;
ALTER TABLE gaming.friendships ADD CONSTRAINT friendships_status_check CHECK (
    status IN ('pending', 'accepted', 'declined')
);
