-- migrate:up
create schema if not exists blog;

create table if not exists blog.posts (
    id uuid primary key default uuidv7(),
    slug text not null unique,
    title text not null,
    content text not null default '',
    content_format text not null default 'tiptap' check (content_format in ('tiptap', 'markdown')),
    status text not null default 'draft' check (status in ('draft', 'published', 'archived')),
    author_id uuid not null references auth.users on delete cascade on update cascade,
    published_at timestamptz,
    featured_image_key text,
    seo_title text,
    seo_description text,
    reading_time_minutes int,
    view_count bigint not null default 0,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp()
);

create index if not exists blog_posts_status_published_at_idx on blog.posts (status, published_at desc);
create index if not exists blog_posts_author_id_idx on blog.posts (author_id);
create index if not exists blog_posts_slug_idx on blog.posts (slug);

create index if not exists blog_posts_fts_idx on blog.posts
    using gin (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(seo_description, '')));

create trigger handle_blog_posts_updated_at before
update on blog.posts for each row execute procedure utility.set_current_timestamp_updated_at();

create table if not exists blog.tags (
    id uuid primary key default uuidv7(),
    name text not null,
    slug text not null unique,
    created_at timestamptz not null default clock_timestamp()
);

create table if not exists blog.post_tags (
    post_id uuid not null references blog.posts on delete cascade on update cascade,
    tag_id uuid not null references blog.tags on delete cascade on update cascade,
    primary key (post_id, tag_id)
);

-- migrate:down
drop table if exists blog.post_tags;
drop table if exists blog.tags;
drop table if exists blog.posts;
drop schema if exists blog;
