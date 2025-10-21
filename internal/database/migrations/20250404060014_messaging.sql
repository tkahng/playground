-- migrate:up
CREATE SCHEMA IF NOT EXISTS messaging;
-- migrate:down
DROP SCHEMA IF EXISTS messaging;