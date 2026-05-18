-- migrate:up
alter table task.workflow_statuses drop constraint if exists workflow_statuses_check;

-- migrate:down
alter table task.workflow_statuses add constraint workflow_statuses_check check (not is_completed or category = 'done');
