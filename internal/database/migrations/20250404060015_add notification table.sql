-- migrate:up
-- Create messaging.notifications table
CREATE TABLE IF NOT EXISTS messaging.notifications (
    id uuid NOT NULL DEFAULT uuidv7(),
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp(),
    read_at timestamptz,
    channel text not null,
    payload jsonb not null default '{}'::jsonb,
    user_id uuid,
    team_member_id uuid,
    team_id uuid,
    metadata jsonb not null default '{}'::jsonb,
    type text not null,
    PRIMARY KEY (id),
    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES auth.users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_notifications_team_member FOREIGN KEY (team_member_id) REFERENCES org.team_members (id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_notifications_team FOREIGN KEY (team_id) REFERENCES org.teams (id) ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create trigger
CREATE TRIGGER handle_messaging_notifications_updated_at before
update on billing.stripe_products for each row execute procedure utility.set_current_timestamp_updated_at();
-- migrate:down
-- Drop trigger
DROP TRIGGER IF EXISTS handle_messaging_notifications_updated_at ON messaging.notifications;
-- Drop table
DROP TABLE IF EXISTS messaging.notifications;