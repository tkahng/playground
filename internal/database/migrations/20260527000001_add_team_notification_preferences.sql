-- migrate:up
CREATE TABLE IF NOT EXISTS messaging.team_notification_preferences (
    id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     timestamptz NOT NULL    DEFAULT clock_timestamp(),
    updated_at     timestamptz NOT NULL    DEFAULT clock_timestamp(),
    team_member_id uuid        NOT NULL    REFERENCES org.team_members (id) ON DELETE CASCADE ON UPDATE CASCADE,
    type           text        NOT NULL,
    enabled        boolean     NOT NULL    DEFAULT true,
    CONSTRAINT uq_team_notification_preferences_member_type UNIQUE (team_member_id, type)
);

CREATE TRIGGER handle_messaging_team_notification_preferences_updated_at
BEFORE UPDATE ON messaging.team_notification_preferences
FOR EACH ROW EXECUTE PROCEDURE utility.set_current_timestamp_updated_at();

CREATE INDEX team_notification_preferences_member_type_enabled_idx
    ON messaging.team_notification_preferences (team_member_id, type, enabled);

-- migrate:down
DROP TABLE IF EXISTS messaging.team_notification_preferences;
