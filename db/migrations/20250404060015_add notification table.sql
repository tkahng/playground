-- migrate:up
-- Create enum type "notification_type"
CREATE TYPE "notification_type" AS ENUM ('mention');
-- Create "notifications" table
CREATE TABLE "notifications" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "channel" text not null default 'general',
    "user_id" uuid NULL,
    "content" jsonb not null default '{}',
    "type" text not null default 'general',
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_notifications_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE
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