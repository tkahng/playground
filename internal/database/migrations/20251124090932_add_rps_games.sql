-- migrate:up
-- rps_games --------------------------------------------------------------------------------
create table gaming.rps_games (
    id uuid primary key default uuidv7(),
    completed_at timestamptz,
    expires_at timestamptz not null,
    status TEXT NOT NULL CHECK (
        status IN ('pending', 'cancelled', 'completed')
    ) default 'pending',
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);
create index idx_gaming_rps_games_completed_at on gaming.rps_games(completed_at);
create index idx_gaming_rps_games_expires_at on gaming.rps_games(expires_at);
create index idx_gaming_rps_games_status on gaming.rps_games(status);
create index idx_gaming_rps_games_metadata_gin on gaming.rps_games using gin(metadata);
create trigger handle_gaming_rps_games_updated_at before
update on gaming.rps_games for each row execute procedure utility.set_current_timestamp_updated_at();
-- rps_participants --------------------------------------------------------------------------------
create table gaming.rps_participants (
    id uuid primary key default uuidv7(),
    game_id uuid not null references gaming.rps_games(id),
    player_id uuid not null references gaming.players(id),
    type TEXT NOT NULL CHECK (type IN ('host', 'guest')) default 'host',
    constraint rps_participants_game_id_player_id_unique unique (game_id, player_id),
    constraint rps_participants_game_id_type_unique unique (game_id, type),
    status TEXT NOT NULL CHECK (
        status IN ('pending', 'declined', 'completed')
    ) default 'pending',
    move TEXT CHECK (move IN ('rock', 'paper', 'scissors')) default 'rock' not null,
    result TEXT CHECK (result IN ('tie', 'win', 'lose')) default 'tie' not null,
    responded_at timestamptz,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);
create index idx_gaming_rps_participants_game_id on gaming.rps_participants(game_id);
create index idx_gaming_rps_participants_player_id on gaming.rps_participants(player_id);
create trigger handle_gaming_rps_participants_updated_at before
update on gaming.rps_participants for each row execute procedure utility.set_current_timestamp_updated_at();
-- migrate:down
-- rps_participants --------------------------------------------------------------------------------
drop trigger handle_gaming_rps_participants_updated_at on gaming.rps_participants;
drop index gaming.idx_gaming_rps_participants_player_id;
drop index gaming.idx_gaming_rps_participants_game_id;
drop table gaming.rps_participants;
-- rps_games --------------------------------------------------------------------------------
drop trigger handle_gaming_rps_games_updated_at on gaming.rps_games;
drop index gaming.idx_gaming_rps_games_metadata_gin;
drop index gaming.idx_gaming_rps_games_status;
drop index gaming.idx_gaming_rps_games_expires_at;
drop index gaming.idx_gaming_rps_games_completed_at;
drop table gaming.rps_games;