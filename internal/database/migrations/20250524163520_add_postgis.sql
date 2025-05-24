-- migrate:up
-- Example: enable the "postgis" extension
create extension if not exists "postgis";
-- migrate:down
drop extension if exists "postgis";