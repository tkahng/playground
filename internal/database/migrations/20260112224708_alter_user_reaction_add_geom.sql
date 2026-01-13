-- migrate:up
alter table sayhello.user_reactions add column geom gis.geometry(Point,4326);

create index if not exists user_reactions_geom_idx on sayhello.user_reactions using gist (geom);

-- migrate:down

drop index if exists user_reactions_geom_idx;

alter table sayhello.user_reactions drop column geom;

