-- migrate:up
create table if not exists storage.media (
    id uuid primary key default uuidv7(),
    user_id uuid references auth.users on delete
    set null on update cascade,
        disk varchar(32) not null,
        directory varchar(255) not null,
        filename varchar(255) not null,
        original_filename varchar(255) not null,
        extension varchar(32) not null,
        mime_type varchar(128) not null,
        size bigint not null,
        created_at timestamptz not null default clock_timestamp(),
        updated_at timestamptz not null default clock_timestamp(),
        constraint media_disk_directory_filename_extension unique(disk, directory, filename, extension)
);
CREATE TRIGGER handle_media_updated_at before
update on storage.media for each row execute procedure utility.set_current_timestamp_updated_at();
-- migrate:down
drop trigger if exists handle_media_updated_at on storage.media;
drop table if exists storage.media;