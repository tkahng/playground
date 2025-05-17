-- migrate:up
CREATE FUNCTION set_current_timestamp_updated_at() RETURNS TRIGGER AS $$
DECLARE _new record;
BEGIN _new := NEW;
_new."updated_at" = clock_timestamp();
RETURN _new;
END;
$$ LANGUAGE plpgsql;
-- Drop old triggers using moddatetime and create new ones using set_current_timestamp_updated_at
-- users
DROP TRIGGER IF EXISTS handle_users_updated_at ON public.users;
CREATE TRIGGER handle_users_updated_at BEFORE
UPDATE ON public.users FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- roles
DROP TRIGGER IF EXISTS handle_roles_updated_at ON public.roles;
CREATE TRIGGER handle_roles_updated_at BEFORE
UPDATE ON public.roles FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- permissions
DROP TRIGGER IF EXISTS handle_permissions_updated_at ON public.permissions;
CREATE TRIGGER handle_permissions_updated_at BEFORE
UPDATE ON public.permissions FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- media
DROP TRIGGER IF EXISTS handle_media_updated_at ON public.media;
CREATE TRIGGER handle_media_updated_at BEFORE
UPDATE ON public.media FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- app_params
DROP TRIGGER IF EXISTS handle_app_params_updated_at ON public.app_params;
CREATE TRIGGER handle_app_params_updated_at BEFORE
UPDATE ON public.app_params FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- tokens
DROP TRIGGER IF EXISTS handle_tokens_updated_at ON public.tokens;
CREATE TRIGGER handle_tokens_updated_at BEFORE
UPDATE ON public.tokens FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- user_accounts
DROP TRIGGER IF EXISTS handle_user_accounts_updated_at ON public.user_accounts;
CREATE TRIGGER handle_user_accounts_updated_at BEFORE
UPDATE ON public.user_accounts FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- user_sessions
DROP TRIGGER IF EXISTS handle_user_sessions_updated_at ON public.user_sessions;
CREATE TRIGGER handle_user_sessions_updated_at BEFORE
UPDATE ON public.user_sessions FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- stripe_customers
DROP TRIGGER IF EXISTS handle_stripe_customers_updated_at ON public.stripe_customers;
CREATE TRIGGER handle_stripe_customers_updated_at BEFORE
UPDATE ON public.stripe_customers FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- stripe_products
DROP TRIGGER IF EXISTS handle_stripe_products_updated_at ON public.stripe_products;
CREATE TRIGGER handle_stripe_products_updated_at BEFORE
UPDATE ON public.stripe_products FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- stripe_prices
DROP TRIGGER IF EXISTS handle_stripe_prices_updated_at ON public.stripe_prices;
CREATE TRIGGER handle_stripe_prices_updated_at BEFORE
UPDATE ON public.stripe_prices FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- stripe_subscriptions
DROP TRIGGER IF EXISTS handle_stripe_subscriptions_updated_at ON public.stripe_subscriptions;
CREATE TRIGGER handle_stripe_subscriptions_updated_at BEFORE
UPDATE ON public.stripe_subscriptions FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- stripe_webhook_events
DROP TRIGGER IF EXISTS handle_stripe_webhook_events_updated_at ON public.stripe_webhook_events;
CREATE TRIGGER handle_stripe_webhook_events_updated_at BEFORE
UPDATE ON public.stripe_webhook_events FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- task_projects
DROP TRIGGER IF EXISTS handle_task_projects_updated_at ON public.task_projects;
CREATE TRIGGER handle_task_projects_updated_at BEFORE
UPDATE ON public.task_projects FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- tasks
DROP TRIGGER IF EXISTS handle_tasks_updated_at ON public.tasks;
CREATE TRIGGER handle_tasks_updated_at BEFORE
UPDATE ON public.tasks FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- migrate:down
-- Revert triggers back to moddatetime
-- users
DROP TRIGGER IF EXISTS handle_users_updated_at ON public.users;
CREATE TRIGGER handle_users_updated_at BEFORE
UPDATE ON public.users FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- roles
DROP TRIGGER IF EXISTS handle_roles_updated_at ON public.roles;
CREATE TRIGGER handle_roles_updated_at BEFORE
UPDATE ON public.roles FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- permissions
DROP TRIGGER IF EXISTS handle_permissions_updated_at ON public.permissions;
CREATE TRIGGER handle_permissions_updated_at BEFORE
UPDATE ON public.permissions FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- media
DROP TRIGGER IF EXISTS handle_media_updated_at ON public.media;
CREATE TRIGGER handle_media_updated_at BEFORE
UPDATE ON public.media FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- app_params
DROP TRIGGER IF EXISTS handle_app_params_updated_at ON public.app_params;
CREATE TRIGGER handle_app_params_updated_at BEFORE
UPDATE ON public.app_params FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- tokens
DROP TRIGGER IF EXISTS handle_tokens_updated_at ON public.tokens;
CREATE TRIGGER handle_tokens_updated_at BEFORE
UPDATE ON public.tokens FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- user_accounts
DROP TRIGGER IF EXISTS handle_user_accounts_updated_at ON public.user_accounts;
CREATE TRIGGER handle_user_accounts_updated_at BEFORE
UPDATE ON public.user_accounts FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- user_sessions
DROP TRIGGER IF EXISTS handle_user_sessions_updated_at ON public.user_sessions;
CREATE TRIGGER handle_user_sessions_updated_at BEFORE
UPDATE ON public.user_sessions FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- stripe_customers
DROP TRIGGER IF EXISTS handle_stripe_customers_updated_at ON public.stripe_customers;
CREATE TRIGGER handle_stripe_customers_updated_at BEFORE
UPDATE ON public.stripe_customers FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- stripe_products
DROP TRIGGER IF EXISTS handle_stripe_products_updated_at ON public.stripe_products;
CREATE TRIGGER handle_stripe_products_updated_at BEFORE
UPDATE ON public.stripe_products FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- stripe_prices
DROP TRIGGER IF EXISTS handle_stripe_prices_updated_at ON public.stripe_prices;
CREATE TRIGGER handle_stripe_prices_updated_at BEFORE
UPDATE ON public.stripe_prices FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- stripe_subscriptions
DROP TRIGGER IF EXISTS handle_stripe_subscriptions_updated_at ON public.stripe_subscriptions;
CREATE TRIGGER handle_stripe_subscriptions_updated_at BEFORE
UPDATE ON public.stripe_subscriptions FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- stripe_webhook_events
DROP TRIGGER IF EXISTS handle_stripe_webhook_events_updated_at ON public.stripe_webhook_events;
CREATE TRIGGER handle_stripe_webhook_events_updated_at BEFORE
UPDATE ON public.stripe_webhook_events FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- task_projects
DROP TRIGGER IF EXISTS handle_task_projects_updated_at ON public.task_projects;
CREATE TRIGGER handle_task_projects_updated_at BEFORE
UPDATE ON public.task_projects FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- tasks
DROP TRIGGER IF EXISTS handle_tasks_updated_at ON public.tasks;
CREATE TRIGGER handle_tasks_updated_at BEFORE
UPDATE ON public.tasks FOR EACH ROW EXECUTE PROCEDURE moddatetime(updated_at);
-- function
DROP FUNCTION set_current_timestamp_updated_at;