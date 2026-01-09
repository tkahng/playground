-- migrate:up
create schema gis;
create extension postgis schema gis;
-- migrate:down
drop schema gis cascade;
