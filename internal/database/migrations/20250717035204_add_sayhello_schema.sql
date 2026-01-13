-- migrate:up
CREATE SCHEMA IF NOT EXISTS sayhello;
-- migrate:down
DROP SCHEMA IF EXISTS sayhello;