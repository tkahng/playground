-- migrate:up
CREATE SCHEMA IF NOT EXISTS task;
-- migrate:down
DROP SCHEMA IF EXISTS task;