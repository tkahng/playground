-- migrate:up
-- projects status ----------------------------------------------------------------------
create type task.task_project_status as enum ('todo', 'in_progress', 'done');
-- project table  ----------------------------------------------------------------------
create table if not exists task.task_projects (
    id uuid primary key default uuidv7(),
    team_id uuid not null references org.teams on delete cascade on update cascade,
    created_by_member_id uuid references org.team_members on delete
    set null on update cascade,
        name text not null,
        description text,
        status task.task_project_status not null default 'todo',
        start_at timestamptz,
        end_at timestamptz,
        assignee_id uuid references org.team_members on delete
    set null on update cascade,
        reporter_id uuid references org.team_members on delete
    set null on update cascade,
        rank double precision not null default 0.0,
        created_at timestamptz not null default clock_timestamp(),
        updated_at timestamptz not null default clock_timestamp()
);
-- project table updated_at trigger  ----------------------------------------------------------------------
create trigger handle_task_projects_updated_at before
update on task.task_projects for each row execute procedure utility.set_current_timestamp_updated_at();
-- tasks status ----------------------------------------------------------------------
create type task.task_status as enum ('todo', 'in_progress', 'done');
-- tasks table  ----------------------------------------------------------------------
create table if not exists task.tasks (
    id uuid primary key default uuidv7(),
    -- user_id uuid not null references public.users on delete cascade on update cascade,
    team_id uuid not null references org.teams on delete cascade on update cascade,
    created_by_member_id uuid references org.team_members on delete
    set null on update cascade,
        project_id uuid not null references task.task_projects on delete cascade on update cascade,
        name text not null,
        description text,
        status task.task_status not null default 'todo',
        start_at timestamptz,
        end_at timestamptz,
        assignee_id uuid references org.team_members on delete
    set null on update cascade,
        reporter_id uuid references org.team_members on delete
    set null on update cascade,
        rank double precision not null default 0.0,
        parent_id uuid references task.tasks on delete
    set null on update cascade,
        created_at timestamptz not null default clock_timestamp(),
        updated_at timestamptz not null default clock_timestamp()
);
-- tasks table updated_at trigger  ----------------------------------------------------------------------
create trigger handle_tasks_updated_at before
update on task.tasks for each row execute procedure utility.set_current_timestamp_updated_at();
-- migrate:down
-- tasks table updated_at trigger  ----------------------------------------------------------------------
drop trigger if exists handle_tasks_updated_at on task.tasks;
-- tasks table  ----------------------------------------------------------------------
drop table if exists task.tasks;
-- tasks status ----------------------------------------------------------------------
drop type if exists task.task_status;
-- project table updated_at trigger  ----------------------------------------------------------------------
drop trigger if exists handle_task_projects_updated_at on task.task_projects;
-- project table  ----------------------------------------------------------------------
drop table if exists task.task_projects;
-- project status ----------------------------------------------------------------------
drop type if exists task.task_project_status;