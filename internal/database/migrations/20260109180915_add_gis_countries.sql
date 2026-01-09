-- migrate:up
create table if not exists gis.countries (
	gid serial primary key,
	name character varying(29) not null,
	iso_a2_eh character varying(5) not null,
	iso_a3_eh character varying(3) not null,
	geom gis.geometry (multipolygon, 4326) not null
);
create index if not exists countries_geom_idx on gis.countries using gist (geom);

-- migrate:down
drop index if exists gis.countries_geom_idx;
drop table if exists gis.countries;
