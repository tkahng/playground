-- migrate:up
-- create utility schema
CREATE SCHEMA IF NOT EXISTS utility;
-- create update timestamp function
CREATE FUNCTION IF NOT EXISTS utility.set_current_timestamp_updated_at() RETURNS TRIGGER AS $$
DECLARE _new record;
BEGIN _new := NEW;
_new."updated_at" = clock_timestamp();
RETURN _new;
END;
$$ LANGUAGE plpgsql;
-- create not_empty check
create or replace function utility.not_empty(input text) returns boolean language plpgsql stable as $$ begin return (char_length(input) > 0);
end;
$$;
-- migrate:down
-- drop not_empty check
drop function if exists utility.not_empty(text);
-- drop update timestamp function
DROP FUNCTION IF EXISTS utility.set_current_timestamp_updated_at();
-- drop utility schema
DROP SCHEMA IF EXISTS utility;