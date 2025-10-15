-- migrate:up
-- create app schema
CREATE SCHEMA IF NOT EXISTS app;
-- migrate:down
-- drop app schema
DROP SCHEMA IF EXISTS app;