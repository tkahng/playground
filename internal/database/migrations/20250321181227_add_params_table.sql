-- migrate:up
create table if not exists app.params (
    id uuid primary key default uuidv7(),
    name text not null unique,
    value jsonb not null,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);
CREATE TRIGGER handle_app_params_updated_at before
update on app.params for each row execute procedure utility.set_current_timestamp_updated_at();
-- migrate:down
drop trigger if exists handle_app_params_updated_at on app.params;
drop table if exists app.params;