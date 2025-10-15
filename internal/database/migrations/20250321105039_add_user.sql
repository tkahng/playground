-------------------------------------------------------------------------------------------------------------------
-- migrate:up
-- create users table
create table if not exists auth.users (
    id uuid not null primary key default uuidv7(),
    email character varying unique not null,
    email_verified_at timestamptz,
    name character varying,
    image text,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);
-- create users table updated_at trigger
CREATE TRIGGER handle_auth_users_updated_at before
update on auth.users for each row execute procedure utility.set_current_timestamp_updated_at();
-------------------------------------------------------------------------------------------------------------------
-- migrate:down
-- drop users table updated_at trigger
drop trigger if exists handle_auth_users_updated_at on auth.users;
-- Drop the users table
DROP TABLE IF EXISTS auth.users;