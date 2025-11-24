-- migrate:up
-- players --------------------------------------------------------------------------------
create table gaming.players (
    id uuid primary key default uuidv7(),
    email text not null unique,
    display_name text,
    user_id uuid references auth.users on delete
    set null on update cascade,
        metadata jsonb not null default '{}'::jsonb,
        created_at timestamptz not null default clock_timestamp(),
        updated_at timestamptz not null default clock_timestamp()
);
create index idx_gaming_players_user_id on gaming.players(user_id);
create index idx_gaming_players_email on gaming.players(email);
create index idx_gaming_players_display_name on gaming.players(display_name);
create index idx_gaming_players_metadata_gin on gaming.players using gin(metadata);
create trigger handle_gaming_players_updated_at before
update on gaming.players for each row execute procedure utility.set_current_timestamp_updated_at();
-- friendships --------------------------------------------------------------------------------
create table gaming.friendships (
    id uuid primary key default uuidv7(),
    requesting_player_id uuid not null references gaming.players(id),
    invited_player_id uuid not null references gaming.players(id),
    status TEXT CHECK (
        status IN ('pending', 'accepted', 'declined')
    ) default 'pending' not null,
    responded_at timestamptz,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);
create index idx_gaming_friendships_requesting_player_id on gaming.friendships(requesting_player_id);
create index idx_gaming_friendships_invited_player_id on gaming.friendships(invited_player_id);
create index idx_gaming_friendships_status on gaming.friendships(status);
create index idx_gaming_friendships_status_invited_player_id_requesting_player_id on gaming.friendships(status, invited_player_id, requesting_player_id);
create index idx_gaming_friendships_responded_at on gaming.friendships(responded_at);
create trigger handle_gaming_friendships_updated_at before
update on gaming.friendships for each row execute procedure utility.set_current_timestamp_updated_at();
-- migrate:down
-- friendships --------------------------------------------------------------------------------
drop trigger handle_gaming_friendships_updated_at on gaming.friendships;
drop index gaming.idx_gaming_friendships_responded_at;
drop index gaming.idx_gaming_friendships_status_invited_player_id_requesting_player_id;
drop index gaming.idx_gaming_friendships_status;
drop index gaming.idx_gaming_friendships_invited_player_id;
drop index gaming.idx_gaming_friendships_requesting_player_id;
drop table gaming.friendships;
-- players --------------------------------------------------------------------------------
drop trigger handle_gaming_players_updated_at on gaming.players;
drop index gaming.idx_gaming_players_metadata_gin;
drop index gaming.idx_gaming_players_display_name;
drop index gaming.idx_gaming_players_email;
drop index gaming.idx_gaming_players_user_id;
drop table gaming.players;