-- migrate:up
create table if not exists task.task_comments (
    id                   uuid        primary key default uuidv7(),
    task_id              uuid        not null references task.tasks(id) on delete cascade on update cascade,
    created_by_member_id uuid        not null references org.team_members(id) on delete cascade on update cascade,
    content              text        not null,
    created_at           timestamptz not null default clock_timestamp(),
    updated_at           timestamptz not null default clock_timestamp()
);

create index if not exists task_comments_task_id_idx on task.task_comments (task_id);

create trigger handle_task_comments_updated_at before
update on task.task_comments for each row execute procedure utility.set_current_timestamp_updated_at();

-- migrate:down
drop table if exists task.task_comments;
