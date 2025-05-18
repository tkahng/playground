-- migrate:up
create table if not exists public.app_params (
    id uuid primary key default gen_random_uuid(),
    name text not null unique,
    value jsonb not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
CREATE TRIGGER handle_app_params_updated_at before
update on public.app_params for each row execute procedure set_current_timestamp_updated_at();
-- migrate:down
drop table if exists public.app_params;
drop trigger if exists handle_app_params_updated_at on public.app_params;