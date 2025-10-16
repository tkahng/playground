-- migrate:up
-- create auth schema
CREATE SCHEMA IF NOT EXISTS billing;
-- migrate:down
-- drop auth schema
DROP SCHEMA IF EXISTS billing;