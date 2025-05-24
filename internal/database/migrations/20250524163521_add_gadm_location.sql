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
-- migrate:down