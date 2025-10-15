-- migrate:up
create table if not exists public.user_reactions (
    id uuid not null primary key default uuidv7(),
    user_id uuid references public.users on delete cascade on update cascade,
    type text not null,
    reaction text,
    ip_address text,
    country text,
    city text,
    metadata jsonb,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);
create trigger handle_user_reactions_updated_at before
update on public.user_reactions for each row execute procedure utility.set_current_timestamp_updated_at();
-- migrate:down
drop trigger if exists handle_user_reactions_updated_at on public.user_reactions;
drop table if exists public.user_reactions;