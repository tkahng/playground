-- migrate:up
-- create table roles
CREATE TABLE if not exists auth.roles (
    id uuid primary key default uuidv7(),
    name varchar(150) not null unique,
    description text,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);
CREATE TRIGGER handle_auth_roles_updated_at before
update on auth.roles for each row execute procedure utility.set_current_timestamp_updated_at();
-- create table permissions
CREATE TABLE if not exists auth.permissions (
    id uuid primary key default uuidv7(),
    name varchar(150) not null unique,
    description text,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);
CREATE TRIGGER handle_auth_permissions_updated_at before
update on auth.permissions for each row execute procedure utility.set_current_timestamp_updated_at();
-- create table user roles
CREATE TABLE if not exists auth.user_roles (
    user_id uuid references auth.users on delete cascade on update cascade not null,
    role_id uuid references auth.roles on delete cascade on update cascade not null,
    primary key (user_id, role_id)
);
-- create table roles permissions
create table if not exists auth.role_permissions (
    role_id uuid references auth.roles on delete cascade on update cascade not null,
    permission_id uuid references auth.permissions on delete cascade on update cascade not null,
    primary key (role_id, permission_id)
);
-- create table user permissions
CREATE TABLE if not exists auth.user_permissions (
    user_id uuid references auth.users on delete cascade on update cascade not null,
    permission_id uuid references auth.permissions on delete cascade on update cascade not null,
    primary key (user_id, permission_id)
);
----------------------------------------------------------------------------------------------------------------------------------------
-- migrate:down
-- drop user permissions
DROP TABLE IF EXISTS auth.user_permissions;
-- drop roles permissions
DROP TABLE IF EXISTS auth.role_permissions;
-- drop user roles
DROP TABLE IF EXISTS auth.user_roles;
-- drop permissions
drop trigger if exists handle_auth_permissions_updated_at on auth.permissions;
DROP TABLE IF EXISTS auth.permissions;
-- drop roles
drop trigger if exists handle_auth_roles_updated_at on auth.roles;
DROP TABLE IF EXISTS auth.roles;