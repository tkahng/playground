-- migrate:up
CREATE SCHEMA IF NOT EXISTS billing;
-- migrate:down
DROP SCHEMA IF EXISTS billing;