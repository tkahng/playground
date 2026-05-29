-- migrate:up
CREATE TABLE app.sse_tickets (
    id          text        NOT NULL PRIMARY KEY,
    user_id     uuid        NOT NULL,
    resource_id uuid        NOT NULL,
    expires_at  timestamptz NOT NULL
);
CREATE INDEX sse_tickets_expires_at_idx ON app.sse_tickets (expires_at);

-- migrate:down
DROP INDEX IF EXISTS sse_tickets_expires_at_idx;
DROP TABLE IF EXISTS app.sse_tickets;
