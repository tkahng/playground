-- migrate:up
--------------- MODDATETIME START -----------------------------------------------------------------------
create extension if not exists moddatetime;


-- migrate:down

drop extension if exists moddatetime;
