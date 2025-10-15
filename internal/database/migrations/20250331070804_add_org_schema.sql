-- migrate:up
-- create auth schema
CREATE SCHEMA IF NOT EXISTS org;
-- migrate:down
-- drop auth schema
DROP SCHEMA IF EXISTS org;