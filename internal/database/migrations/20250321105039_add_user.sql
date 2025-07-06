-- migrate:up
--------------- USER TABLE START -----------------------------------------------------------------------
create table if not exists public.users (
    id uuid not null primary key default gen_random_uuid(),
    email character varying unique not null,
    email_verified_at timestamptz,
    name character varying,
    image text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
-- this trigger will set the "updated_at" column to the current timestamptz for every update
CREATE TRIGGER handle_users_updated_at before
update on public.users for each row execute procedure set_current_timestamp_updated_at();
-- migrate:down
drop trigger if exists handle_users_updated_at on public.users;
-- Drop the users table
DROP TABLE IF EXISTS public.users;