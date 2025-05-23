-- migrate:up
-- Create "notifications" table
CREATE TABLE "notifications" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "read_at" timestamptz,
    "channel" text not null,
    "user_id" uuid,
    "team_member_id" uuid,
    "team_id" uuid,
    "metadata" jsonb not null default '{}',
    "type" text not null,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_notifications_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT "fk_notifications_team_member" FOREIGN KEY ("team_member_id") REFERENCES "team_members" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT "fk_notifications_team" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Add function "notify_after_notification_insert"
CREATE OR REPLACE FUNCTION notify_after_notification_insert() RETURNS TRIGGER AS $BODY$ BEGIN PERFORM pg_notify('notification', row_to_json(NEW)::text);
PERFORM pg_notify(NEW.channel, row_to_json(NEW)::text);
RETURN NEW;
END;
$BODY$ LANGUAGE plpgsql;
-- Add trigger "trigger_notify_after_notification_insert"
CREATE OR REPLACE TRIGGER trigger_notify_after_notification_insert
AFTER
INSERT ON notifications FOR EACH ROW EXECUTE PROCEDURE notify_after_notification_insert();
-- migrate:down
DROP TRIGGER IF EXISTS trigger_notify_after_notification_insert ON notifications;
DROP FUNCTION IF EXISTS notify_after_notification_insert;
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_type;