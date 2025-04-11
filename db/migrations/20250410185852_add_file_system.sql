-- migrate:up
create table if not exists public.media (
    id uuid primary key default gen_random_uuid(),
    user_id uuid references public.users on delete
    set null on update cascade,
        disk varchar(32) not null,
        directory varchar(255) not null,
        filename varchar(255) not null,
        original_filename varchar(255) not null,
        extension varchar(32) not null,
        mime_type varchar(128) not null,
        size bigint not null,
        created_at timestamp with time zone not null default now(),
        updated_at timestamp with time zone not null default now(),
        constraint media_disk_directory_filename_extension unique(disk, directory, filename, extension)
);
CREATE TRIGGER handle_media_updated_at before
update on public.media for each row execute procedure moddatetime(updated_at);
-- migrate:down
drop trigger if exists handle_media_updated_at on public.media;
drop table if exists public.media;