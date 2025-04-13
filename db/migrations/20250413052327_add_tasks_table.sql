-- migrate:up
------------------------------
------------------------------
------------------------------
-- projects status ----------------------------------------------------------------------
create type "task_project_status" as enum ('todo', 'in_progress', 'done');
-- project table  ----------------------------------------------------------------------
create table if not exists public.projects (
    id uuid primary key default gen_random_uuid(),
    name text not null unique,
    description text,
    order integer not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
-- project table updated_at trigger  ----------------------------------------------------------------------
create trigger handle_projects_updated_at before
update on public.projects for each row execute procedure moddatetime(updated_at);
-- tasks status ----------------------------------------------------------------------
create type "task_status" as enum ('todo', 'in_progress', 'done');
-- tasks table  ----------------------------------------------------------------------
create table if not exists public.tasks (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    description text,
    status task_status not null default 'todo',
    order integer not null default 0,
    parent_id uuid references public.tasks on delete cascade on update cascade,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
-- tasks table updated_at trigger  ----------------------------------------------------------------------
create trigger handle_tasks_updated_at before
update on public.tasks for each row execute procedure moddatetime(updated_at);
------------------------------
------------------------------
------------------------------
-- migrate:down
------------------------------
------------------------------
------------------------------
-- tasks table updated_at trigger  ----------------------------------------------------------------------
drop trigger if exists handle_tasks_updated_at on public.tasks;
-- tasks table  ----------------------------------------------------------------------
drop table if exists public.tasks;
-- tasks status ----------------------------------------------------------------------
drop type if exists "task_status";
-- project table updated_at trigger  ----------------------------------------------------------------------
drop trigger if exists handle_projects_updated_at on public.projects;
-- project table  ----------------------------------------------------------------------
drop table if exists public.projects;
-- project status ----------------------------------------------------------------------
drop type if exists "task_project_status";