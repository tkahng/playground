-- migrate:up
CREATE SCHEMA IF NOT EXISTS storage;
-- migrate:down
DROP SCHEMA IF EXISTS storage;