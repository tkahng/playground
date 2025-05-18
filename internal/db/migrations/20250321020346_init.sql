-- migrate:up
CREATE FUNCTION set_current_timestamp_updated_at() RETURNS TRIGGER AS $$
DECLARE _new record;
BEGIN _new := NEW;
_new."updated_at" = clock_timestamp();
RETURN _new;
END;
$$ LANGUAGE plpgsql;
-- migrate:down
DROP FUNCTION IF EXISTS set_current_timestamp_updated_at();