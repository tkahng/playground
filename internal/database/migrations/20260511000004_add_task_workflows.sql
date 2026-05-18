-- migrate:up
-- Workflows move project/task status away from hard-coded todo/in_progress/done
-- enums and toward a Trello/Jira-style model where teams can define their own
-- columns/statuses. This migration is intentionally additive: the existing enum
-- status columns remain in place while the application learns to dual-read and
-- dual-write workflow_status_id values in follow-up migrations.
--
-- Migration intent: once the enum columns are removed, a future down migration
-- should not try to reconstruct enum status values from arbitrary configurable
-- workflows. That semantic reversal is lossy once teams can rename statuses,
-- reorder columns, or add statuses that do not map cleanly to the legacy enum.
-- If rollback is needed after that cutover, prefer restoring from backup or a
-- purpose-built forward migration over pretending the data can be made perfectly
-- consistent with todo/in_progress/done again.
create table if not exists task.workflows (
    id uuid primary key default uuidv7(),
    team_id uuid not null references org.teams on delete cascade on update cascade,
    created_by_member_id uuid references org.team_members on delete set null on update cascade,
    applies_to text not null check (applies_to in ('project', 'task')),
    name text not null,
    description text,
    is_default boolean not null default false,
    is_archived boolean not null default false,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp(),
    unique (team_id, applies_to, name)
);

create unique index if not exists workflows_one_default_per_team_target
    on task.workflows (team_id, applies_to)
    where is_default and not is_archived;

create trigger handle_workflows_updated_at before
update on task.workflows for each row execute procedure utility.set_current_timestamp_updated_at();

create table if not exists task.workflow_statuses (
    id uuid primary key default uuidv7(),
    workflow_id uuid not null references task.workflows on delete cascade on update cascade,
    name text not null,
    slug text not null,
    description text,
    category text not null check (category in ('todo', 'in_progress', 'done')),
    color text,
    rank double precision not null default 0.0,
    is_completed boolean not null default false,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp(),
    unique (workflow_id, slug)
);

create index if not exists workflow_statuses_workflow_rank_idx
    on task.workflow_statuses (workflow_id, rank);

create trigger handle_workflow_statuses_updated_at before
update on task.workflow_statuses for each row execute procedure utility.set_current_timestamp_updated_at();

insert into task.workflows (team_id, applies_to, name, description, is_default)
select teams.id,
    target.applies_to,
    target.name,
    target.description,
    true
from org.teams
cross join (
    values
        ('project', 'Default project workflow', 'Default workflow for project lifecycle status.'),
        ('task', 'Default task workflow', 'Default workflow for task board status.')
) as target(applies_to, name, description)
on conflict (team_id, applies_to, name) do nothing;

insert into task.workflow_statuses (workflow_id, name, slug, category, color, rank, is_completed)
select workflows.id,
    status.name,
    status.slug,
    status.category,
    status.color,
    status.rank,
    status.is_completed
from task.workflows
cross join (
    values
        ('To do', 'todo', 'todo', '#6b7280', 1000.0, false),
        ('In progress', 'in_progress', 'in_progress', '#2563eb', 2000.0, false),
        ('Done', 'done', 'done', '#16a34a', 3000.0, true)
) as status(name, slug, category, color, rank, is_completed)
where workflows.applies_to in ('project', 'task')
on conflict (workflow_id, slug) do nothing;

alter table task.task_projects
    add column if not exists workflow_id uuid references task.workflows on delete restrict on update cascade,
    add column if not exists workflow_status_id uuid references task.workflow_statuses on delete set null on update cascade;

alter table task.tasks
    add column if not exists workflow_status_id uuid references task.workflow_statuses on delete set null on update cascade;

update task.task_projects
set workflow_id = workflows.id
from task.workflows
where workflows.team_id = task_projects.team_id
  and workflows.applies_to = 'task'
  and workflows.is_default
  and task_projects.workflow_id is null;

update task.task_projects
set workflow_status_id = workflow_statuses.id
from task.workflows
join task.workflow_statuses on workflow_statuses.workflow_id = workflows.id
where workflows.team_id = task_projects.team_id
  and workflows.applies_to = 'project'
  and workflows.is_default
  and workflow_statuses.slug = task_projects.status::text
  and task_projects.workflow_status_id is null;

update task.tasks
set workflow_status_id = workflow_statuses.id
from task.task_projects
join task.workflow_statuses on workflow_statuses.workflow_id = task_projects.workflow_id
where task_projects.id = tasks.project_id
  and workflow_statuses.slug = tasks.status::text
  and tasks.workflow_status_id is null;

create index if not exists task_projects_workflow_id_idx
    on task.task_projects (workflow_id);

create index if not exists task_projects_workflow_status_id_idx
    on task.task_projects (workflow_status_id);

create index if not exists tasks_workflow_status_id_idx
    on task.tasks (workflow_status_id);

-- Cross-team guard: prevent workflow_id / workflow_status_id referencing a
-- different team's workflow. The FK alone cannot enforce this because it only
-- checks row existence, not team ownership.
create or replace function task.check_project_workflow_team()
returns trigger as $$
begin
    if new.workflow_id is not null then
        if not exists (
            select 1 from task.workflows
            where id = new.workflow_id and team_id = new.team_id
        ) then
            raise exception 'workflow_id % does not belong to team %', new.workflow_id, new.team_id;
        end if;
    end if;
    if new.workflow_status_id is not null then
        if not exists (
            select 1 from task.workflow_statuses ws
            join task.workflows w on w.id = ws.workflow_id
            where ws.id = new.workflow_status_id and w.team_id = new.team_id
        ) then
            raise exception 'workflow_status_id % does not belong to team %', new.workflow_status_id, new.team_id;
        end if;
    end if;
    return new;
end;
$$ language plpgsql;

create trigger check_task_projects_workflow_team
    before insert or update on task.task_projects
    for each row execute procedure task.check_project_workflow_team();

create or replace function task.check_task_workflow_status_team()
returns trigger as $$
begin
    if new.workflow_status_id is not null then
        if not exists (
            select 1 from task.workflow_statuses ws
            join task.workflows w on w.id = ws.workflow_id
            where ws.id = new.workflow_status_id and w.team_id = new.team_id
        ) then
            raise exception 'workflow_status_id % does not belong to team %', new.workflow_status_id, new.team_id;
        end if;
    end if;
    return new;
end;
$$ language plpgsql;

create trigger check_tasks_workflow_status_team
    before insert or update on task.tasks
    for each row execute procedure task.check_task_workflow_status_team();

-- Verify the backfill above was complete. Fails the migration (rolls back) if
-- any pre-existing row was not reached by the three UPDATE statements above.
do $$
declare
    unset_projects int;
    unset_tasks    int;
begin
    select count(*) into unset_projects
    from task.task_projects
    where workflow_id is null or workflow_status_id is null;

    select count(*) into unset_tasks
    from task.tasks
    where workflow_status_id is null;

    if unset_projects > 0 then
        raise exception 'workflow backfill incomplete: % project row(s) missing workflow_id or workflow_status_id',
            unset_projects;
    end if;
    if unset_tasks > 0 then
        raise exception 'workflow backfill incomplete: % task row(s) missing workflow_status_id',
            unset_tasks;
    end if;
end $$;

-- migrate:down
-- This down migration removes the workflow foundation added above. It does not
-- attempt to preserve configurable workflow definitions anywhere else. That is
-- acceptable at this additive stage because the legacy enum status columns still
-- exist; it is also the documented direction for the later enum-removal cutover,
-- where perfect down-migration data consistency is intentionally out of scope.
drop index if exists task.tasks_workflow_status_id_idx;
drop index if exists task.task_projects_workflow_status_id_idx;
drop index if exists task.task_projects_workflow_id_idx;

drop trigger if exists check_tasks_workflow_status_team on task.tasks;
drop function if exists task.check_task_workflow_status_team();

alter table task.tasks
    drop column if exists workflow_status_id;

drop trigger if exists check_task_projects_workflow_team on task.task_projects;
drop function if exists task.check_project_workflow_team();

alter table task.task_projects
    drop column if exists workflow_status_id,
    drop column if exists workflow_id;

drop trigger if exists handle_workflow_statuses_updated_at on task.workflow_statuses;
drop table if exists task.workflow_statuses;

drop trigger if exists handle_workflows_updated_at on task.workflows;
drop table if exists task.workflows;
