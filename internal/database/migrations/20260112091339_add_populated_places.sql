-- migrate:up
CREATE TABLE IF NOT EXISTS gis.populated_places
(
    gid serial primary key, 
    geom gis.geometry(Point,4326) not null, 
    scalerank numeric(2,0) NOT NULL CHECK (scalerank >= 0 AND scalerank <= 10),
    labelrank numeric(2,0) NOT NULL CHECK (labelrank >= 0 AND labelrank <= 10),
    featurecla character varying(50) not null,
    name character varying(100) not null, 
    nameascii character varying(100) not null, 
    sov0name character varying(100) not null,
    adm0name character varying(50) not null, 
    adm0_a3 character varying(3) not null, 
    adm1name character varying(100) not null, -- filter out nullable
    iso_a2 character varying(5) not null, -- filter out -99
    pop_max numeric(12,0) NOT NULL CHECK (pop_max >= 0), -- filter -99
    min_zoom numeric(2,1) not null
);

-- Primary spatial index
CREATE INDEX IF NOT EXISTS populated_places_geom_idx 
ON gis.populated_places USING GIST (geom);

-- Composite for relevance-ordered filters (scalerank + pop_max DESC)
CREATE INDEX IF NOT EXISTS populated_places_scalerank_pop_idx 
ON gis.populated_places (scalerank, pop_max DESC);

-- Partial GIST for major cities (fast coarse queries)
CREATE INDEX IF NOT EXISTS populated_places_geom_scalerank_idx 
ON gis.populated_places USING GIST (geom) 
WHERE scalerank <= 4;

-- migrate:down
DROP INDEX IF EXISTS populated_places_geom_scalerank_idx;
DROP INDEX IF EXISTS populated_places_scalerank_pop_idx;
DROP INDEX IF EXISTS populated_places_geom_idx;
DROP TABLE IF EXISTS gis.populated_places;
