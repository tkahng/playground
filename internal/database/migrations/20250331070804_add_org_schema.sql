-- migrate:up
CREATE SCHEMA IF NOT EXISTS org;
-- migrate:down
DROP SCHEMA IF EXISTS org;