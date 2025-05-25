-- migrate:up
CREATE TABLE public.gadm_boundaries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    iso3 TEXT,
    name_0 TEXT,
    name_1 TEXT,
    name_2 TEXT,
    name_3 TEXT,
    level INT,
    geom GEOMETRY(MULTIPOLYGON, 4326)
);
CREATE INDEX if not exists gadm_boundaries_gist ON public.gadm_boundaries USING GIST (geom);
-- migrate:down
DROP INDEX if exists gadm_boundaries_gist;
DROP TABLE public.gadm_boundaries;