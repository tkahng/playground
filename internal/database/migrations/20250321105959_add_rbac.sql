-- migrate:up
-- roles
CREATE TABLE if not exists public.roles (
    id uuid primary key default gen_random_uuid(),
    -- id uuid primary key default gen_random_uuid(),
    name varchar(150) not null unique,
    -- name of the role. e.g. admin, user
    -- code varchar(100) not null unique,
    -- code of the role. "role:admin", "role:user"
    description text,
    -- description of the role
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
CREATE TRIGGER handle_roles_updated_at before
update on public.roles for each row execute procedure set_current_timestamp_updated_at();
-- permissions
CREATE TABLE if not exists public.permissions (
    id uuid primary key default gen_random_uuid(),
    -- id uuid primary key default gen_random_uuid(),
    name varchar(150) not null unique,
    -- name of the permission in ${action}:${resource}. e.g. read:users, user
    -- code varchar(100) not null unique,
    -- code of the role. "role:admin", "role:user"
    description text,
    -- description of the role
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
CREATE TRIGGER handle_permissions_updated_at before
update on public.permissions for each row execute procedure set_current_timestamp_updated_at();
-- user roles
CREATE TABLE if not exists public.user_roles (
    user_id uuid references public.users on delete cascade on update cascade not null,
    role_id uuid references public.roles on delete cascade on update cascade not null,
    primary key (user_id, role_id)
);
-- roles permissions
create table if not exists public.role_permissions (
    role_id uuid references public.roles on delete cascade on update cascade not null,
    permission_id uuid references public.permissions on delete cascade on update cascade not null,
    primary key (role_id, permission_id)
);
-- user permissions
CREATE TABLE if not exists public.user_permissions (
    user_id uuid references public.users on delete cascade on update cascade not null,
    permission_id uuid references public.permissions on delete cascade on update cascade not null,
    primary key (user_id, permission_id)
);
----------------------------------------------------------------------------------------------------------------------------------------
-- migrate:down
-- user permissions
DROP TABLE IF EXISTS public.user_permissions;
-- roles permissions
DROP TABLE IF EXISTS public.role_permissions;
-- user roles
DROP TABLE IF EXISTS public.user_roles;
-- permissions
drop trigger if exists handle_permissions_updated_at on public.permissions;
DROP TABLE IF EXISTS public.permissions;
-- roles
drop trigger if exists handle_roles_updated_at on public.roles;
DROP TABLE IF EXISTS public.roles;