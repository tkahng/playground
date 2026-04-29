-- migrate:up
-- Fix trigger: was incorrectly targeting billing.stripe_products
DROP TRIGGER IF EXISTS handle_messaging_notifications_updated_at ON billing.stripe_products;
CREATE TRIGGER handle_messaging_notifications_updated_at BEFORE
UPDATE ON messaging.notifications FOR EACH ROW EXECUTE PROCEDURE utility.set_current_timestamp_updated_at();

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS notifications_team_member_read_at_idx ON messaging.notifications (team_member_id, read_at);
CREATE INDEX IF NOT EXISTS notifications_created_at_idx ON messaging.notifications (created_at DESC);

-- migrate:down
DROP INDEX IF EXISTS notifications_created_at_idx;
DROP INDEX IF EXISTS notifications_team_member_read_at_idx;
DROP TRIGGER IF EXISTS handle_messaging_notifications_updated_at ON messaging.notifications;
