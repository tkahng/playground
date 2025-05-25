SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: job_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.job_status AS ENUM (
    'pending',
    'processing',
    'done',
    'failed'
);


--
-- Name: provider_types; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.provider_types AS ENUM (
    'oauth',
    'credentials'
);


--
-- Name: providers; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.providers AS ENUM (
    'google',
    'apple',
    'facebook',
    'github',
    'credentials'
);


--
-- Name: stripe_customer_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.stripe_customer_type AS ENUM (
    'user',
    'team'
);


--
-- Name: stripe_pricing_plan_interval; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.stripe_pricing_plan_interval AS ENUM (
    'day',
    'week',
    'month',
    'year'
);


--
-- Name: stripe_pricing_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.stripe_pricing_type AS ENUM (
    'one_time',
    'recurring'
);


--
-- Name: stripe_subscription_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.stripe_subscription_status AS ENUM (
    'trialing',
    'active',
    'canceled',
    'incomplete',
    'incomplete_expired',
    'past_due',
    'unpaid',
    'paused'
);


--
-- Name: task_project_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.task_project_status AS ENUM (
    'todo',
    'in_progress',
    'done'
);


--
-- Name: task_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.task_status AS ENUM (
    'todo',
    'in_progress',
    'done'
);


--
-- Name: team_invitation_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.team_invitation_status AS ENUM (
    'pending',
    'accepted',
    'declined',
    'cancelled'
);


--
-- Name: team_member_role; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.team_member_role AS ENUM (
    'owner',
    'member',
    'guest'
);


--
-- Name: token_types; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.token_types AS ENUM (
    'access_token',
    'recovery_token',
    'invite_token',
    'team_invite_token',
    'reauthentication_token',
    'refresh_token',
    'verification_token',
    'password_reset_token',
    'state_token'
);


--
-- Name: not_empty(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.not_empty(input text) RETURNS boolean
    LANGUAGE plpgsql STABLE
    AS $$ begin return (char_length(input) > 0);
end;
$$;


--
-- Name: notify_after_notification_insert(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.notify_after_notification_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN PERFORM pg_notify('notification', row_to_json(NEW)::text);
PERFORM pg_notify(NEW.channel, row_to_json(NEW)::text);
RETURN NEW;
END;
$$;


--
-- Name: set_current_timestamp_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_current_timestamp_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE _new record;
BEGIN _new := NEW;
_new."updated_at" = clock_timestamp();
RETURN _new;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: ai_usages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_usages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    prompt_tokens bigint NOT NULL,
    completion_tokens bigint NOT NULL,
    total_tokens bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: app_params; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.app_params (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    value jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    kind text NOT NULL,
    unique_key text,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    status public.job_status DEFAULT 'pending'::public.job_status NOT NULL,
    run_after timestamp with time zone DEFAULT now() NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 3 NOT NULL,
    last_error text,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    level integer DEFAULT 0 NOT NULL,
    source text,
    message text NOT NULL,
    data jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: media; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.media (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    disk character varying(32) NOT NULL,
    directory character varying(255) NOT NULL,
    filename character varying(255) NOT NULL,
    original_filename character varying(255) NOT NULL,
    extension character varying(32) NOT NULL,
    mime_type character varying(128) NOT NULL,
    size bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    read_at timestamp with time zone,
    channel text NOT NULL,
    user_id uuid,
    team_member_id uuid,
    team_id uuid,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    type text NOT NULL
);


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(150) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: product_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product_permissions (
    product_id text NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: product_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product_roles (
    product_id text NOT NULL,
    role_id uuid NOT NULL
);


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_permissions (
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(150) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(128) NOT NULL
);


--
-- Name: stripe_customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stripe_customers (
    id text NOT NULL,
    email text NOT NULL,
    name text,
    customer_type public.stripe_customer_type NOT NULL,
    user_id uuid,
    team_id uuid,
    billing_address jsonb,
    payment_method jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT stripe_customers_only_one_reference_check CHECK ((((user_id IS NOT NULL) AND (customer_type = 'user'::public.stripe_customer_type) AND (team_id IS NULL)) OR ((user_id IS NULL) AND (team_id IS NOT NULL) AND (customer_type = 'team'::public.stripe_customer_type))))
);


--
-- Name: stripe_prices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stripe_prices (
    id text NOT NULL,
    product_id text NOT NULL,
    lookup_key text,
    active boolean DEFAULT false NOT NULL,
    unit_amount bigint,
    currency text NOT NULL,
    type public.stripe_pricing_type NOT NULL,
    "interval" public.stripe_pricing_plan_interval,
    interval_count bigint,
    trial_period_days bigint,
    metadata jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT stripe_prices_currency_check CHECK ((char_length(currency) = 3))
);


--
-- Name: stripe_products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stripe_products (
    id text NOT NULL,
    active boolean DEFAULT false NOT NULL,
    name text NOT NULL,
    description text,
    image text,
    metadata jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: stripe_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stripe_subscriptions (
    id text NOT NULL,
    stripe_customer_id text NOT NULL,
    status public.stripe_subscription_status NOT NULL,
    metadata jsonb NOT NULL,
    item_id text NOT NULL,
    price_id text NOT NULL,
    quantity bigint NOT NULL,
    cancel_at_period_end boolean DEFAULT false NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    current_period_start timestamp with time zone DEFAULT now() NOT NULL,
    current_period_end timestamp with time zone DEFAULT now() NOT NULL,
    ended_at timestamp with time zone,
    cancel_at timestamp with time zone,
    canceled_at timestamp with time zone,
    trial_start timestamp with time zone,
    trial_end timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: stripe_webhook_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stripe_webhook_events (
    id text NOT NULL,
    type text NOT NULL,
    object_type text NOT NULL,
    object_stripe_id text NOT NULL,
    event_creation_date timestamp with time zone NOT NULL,
    request_id text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: task_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_projects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    team_id uuid NOT NULL,
    created_by uuid NOT NULL,
    name text NOT NULL,
    description text,
    status public.task_project_status DEFAULT 'todo'::public.task_project_status NOT NULL,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    assignee_id uuid,
    assigner_id uuid,
    "order" double precision DEFAULT 0.0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    team_id uuid NOT NULL,
    created_by uuid NOT NULL,
    project_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    status public.task_status DEFAULT 'todo'::public.task_status NOT NULL,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    assignee_id uuid,
    assigner_id uuid,
    "order" double precision DEFAULT 0.0 NOT NULL,
    parent_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: team_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.team_invitations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    team_id uuid NOT NULL,
    inviter_member_id uuid NOT NULL,
    email text NOT NULL,
    role public.team_member_role NOT NULL,
    token text NOT NULL,
    status public.team_invitation_status DEFAULT 'pending'::public.team_invitation_status NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: team_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.team_members (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    team_id uuid NOT NULL,
    user_id uuid,
    active boolean DEFAULT true NOT NULL,
    role public.team_member_role NOT NULL,
    has_billing_access boolean DEFAULT false NOT NULL,
    last_selected_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teams (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    type public.token_types NOT NULL,
    user_id uuid,
    otp character varying(255),
    identifier text NOT NULL,
    expires timestamp with time zone NOT NULL,
    token text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT tokens_type_identifier_token_not_empty CHECK ((public.not_empty(identifier) AND public.not_empty(token)))
);


--
-- Name: user_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    type public.provider_types NOT NULL,
    provider public.providers NOT NULL,
    provider_account_id character varying(255) NOT NULL,
    password text,
    refresh_token text,
    access_token text,
    expires_at bigint,
    id_token text,
    scope text,
    session_state text,
    token_type text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: user_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_permissions (
    user_id uuid NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_roles (
    user_id uuid NOT NULL,
    role_id uuid NOT NULL
);


--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    expires timestamp with time zone NOT NULL,
    session_token character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_sessions_token_not_empty CHECK (public.not_empty((session_token)::text))
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying NOT NULL,
    email_verified_at timestamp with time zone,
    name character varying,
    image text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: ai_usages ai_usages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_usages
    ADD CONSTRAINT ai_usages_pkey PRIMARY KEY (id);


--
-- Name: app_params app_params_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_params
    ADD CONSTRAINT app_params_name_key UNIQUE (name);


--
-- Name: app_params app_params_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_params
    ADD CONSTRAINT app_params_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- Name: logs logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);


--
-- Name: media media_disk_directory_filename_extension; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media
    ADD CONSTRAINT media_disk_directory_filename_extension UNIQUE (disk, directory, filename, extension);


--
-- Name: media media_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media
    ADD CONSTRAINT media_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_name_key UNIQUE (name);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: product_permissions product_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_permissions
    ADD CONSTRAINT product_permissions_pkey PRIMARY KEY (product_id, permission_id);


--
-- Name: product_roles product_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_roles
    ADD CONSTRAINT product_roles_pkey PRIMARY KEY (product_id, role_id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: stripe_customers stripe_customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_customers
    ADD CONSTRAINT stripe_customers_pkey PRIMARY KEY (id);


--
-- Name: stripe_customers stripe_customers_team_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_customers
    ADD CONSTRAINT stripe_customers_team_id_key UNIQUE (team_id);


--
-- Name: stripe_customers stripe_customers_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_customers
    ADD CONSTRAINT stripe_customers_user_id_key UNIQUE (user_id);


--
-- Name: stripe_prices stripe_prices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_prices
    ADD CONSTRAINT stripe_prices_pkey PRIMARY KEY (id);


--
-- Name: stripe_products stripe_products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_products
    ADD CONSTRAINT stripe_products_pkey PRIMARY KEY (id);


--
-- Name: stripe_subscriptions stripe_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_subscriptions
    ADD CONSTRAINT stripe_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: stripe_webhook_events stripe_webhook_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_webhook_events
    ADD CONSTRAINT stripe_webhook_events_pkey PRIMARY KEY (id);


--
-- Name: task_projects task_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_projects
    ADD CONSTRAINT task_projects_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: team_invitations team_invitations_email_team_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_invitations
    ADD CONSTRAINT team_invitations_email_team_id UNIQUE (email, team_id);


--
-- Name: team_invitations team_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_invitations
    ADD CONSTRAINT team_invitations_pkey PRIMARY KEY (id);


--
-- Name: team_invitations team_invitations_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_invitations
    ADD CONSTRAINT team_invitations_token_key UNIQUE (token);


--
-- Name: team_members team_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_pkey PRIMARY KEY (id);


--
-- Name: team_members team_members_user_id_team_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_user_id_team_id UNIQUE (user_id, team_id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: teams teams_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_slug_key UNIQUE (slug);


--
-- Name: tokens tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tokens
    ADD CONSTRAINT tokens_pkey PRIMARY KEY (id);


--
-- Name: tokens tokens_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tokens
    ADD CONSTRAINT tokens_token_key UNIQUE (token);


--
-- Name: user_accounts user_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_accounts
    ADD CONSTRAINT user_accounts_pkey PRIMARY KEY (id);


--
-- Name: user_accounts user_accounts_provider_provider_account_id_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_accounts
    ADD CONSTRAINT user_accounts_provider_provider_account_id_unique UNIQUE (provider, provider_account_id);


--
-- Name: user_accounts user_accounts_user_id_provider_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_accounts
    ADD CONSTRAINT user_accounts_user_id_provider_unique UNIQUE (user_id, provider);


--
-- Name: user_permissions user_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_permissions
    ADD CONSTRAINT user_permissions_pkey PRIMARY KEY (user_id, permission_id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_session_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_session_token_key UNIQUE (session_token);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_logs_created_at ON public.logs USING btree (created_at);


--
-- Name: idx_logs_data_gin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_logs_data_gin ON public.logs USING gin (data);


--
-- Name: idx_logs_level; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_logs_level ON public.logs USING btree (level);


--
-- Name: idx_logs_source; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_logs_source ON public.logs USING btree (source);


--
-- Name: jobs_polling_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX jobs_polling_idx ON public.jobs USING btree (status, run_after, attempts);


--
-- Name: uniq_jobs_active_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_jobs_active_key ON public.jobs USING btree (unique_key) WHERE (status = ANY (ARRAY['pending'::public.job_status, 'processing'::public.job_status]));


--
-- Name: app_params handle_app_params_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_app_params_updated_at BEFORE UPDATE ON public.app_params FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: media handle_media_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_media_updated_at BEFORE UPDATE ON public.media FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: permissions handle_permissions_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_permissions_updated_at BEFORE UPDATE ON public.permissions FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: roles handle_roles_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_roles_updated_at BEFORE UPDATE ON public.roles FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: stripe_customers handle_stripe_customers_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_stripe_customers_updated_at BEFORE UPDATE ON public.stripe_customers FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: stripe_prices handle_stripe_prices_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_stripe_prices_updated_at BEFORE UPDATE ON public.stripe_prices FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: stripe_products handle_stripe_products_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_stripe_products_updated_at BEFORE UPDATE ON public.stripe_products FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: stripe_subscriptions handle_stripe_subscriptions_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_stripe_subscriptions_updated_at BEFORE UPDATE ON public.stripe_subscriptions FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: stripe_webhook_events handle_stripe_webhook_events_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_stripe_webhook_events_updated_at BEFORE UPDATE ON public.stripe_webhook_events FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: task_projects handle_task_projects_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_task_projects_updated_at BEFORE UPDATE ON public.task_projects FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: tasks handle_tasks_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_tasks_updated_at BEFORE UPDATE ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: team_members handle_team_members_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_team_members_updated_at BEFORE UPDATE ON public.team_members FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: teams handle_teams_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_teams_updated_at BEFORE UPDATE ON public.teams FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: tokens handle_tokens_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_tokens_updated_at BEFORE UPDATE ON public.tokens FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: user_accounts handle_user_accounts_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_user_accounts_updated_at BEFORE UPDATE ON public.user_accounts FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: user_sessions handle_user_sessions_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_user_sessions_updated_at BEFORE UPDATE ON public.user_sessions FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: users handle_users_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER handle_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();


--
-- Name: notifications trigger_notify_after_notification_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_notify_after_notification_insert AFTER INSERT ON public.notifications FOR EACH ROW EXECUTE FUNCTION public.notify_after_notification_insert();


--
-- Name: ai_usages ai_usages_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_usages
    ADD CONSTRAINT ai_usages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notifications fk_notifications_team; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_team FOREIGN KEY (team_id) REFERENCES public.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notifications fk_notifications_team_member; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_team_member FOREIGN KEY (team_member_id) REFERENCES public.team_members(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notifications fk_notifications_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: media media_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media
    ADD CONSTRAINT media_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: product_permissions product_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_permissions
    ADD CONSTRAINT product_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product_permissions product_permissions_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_permissions
    ADD CONSTRAINT product_permissions_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.stripe_products(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product_roles product_roles_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_roles
    ADD CONSTRAINT product_roles_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.stripe_products(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product_roles product_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_roles
    ADD CONSTRAINT product_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stripe_customers stripe_customers_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_customers
    ADD CONSTRAINT stripe_customers_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stripe_customers stripe_customers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_customers
    ADD CONSTRAINT stripe_customers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stripe_prices stripe_prices_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_prices
    ADD CONSTRAINT stripe_prices_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.stripe_products(id);


--
-- Name: stripe_subscriptions stripe_subscriptions_price_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_subscriptions
    ADD CONSTRAINT stripe_subscriptions_price_id_fkey FOREIGN KEY (price_id) REFERENCES public.stripe_prices(id);


--
-- Name: stripe_subscriptions stripe_subscriptions_stripe_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stripe_subscriptions
    ADD CONSTRAINT stripe_subscriptions_stripe_customer_id_fkey FOREIGN KEY (stripe_customer_id) REFERENCES public.stripe_customers(id);


--
-- Name: task_projects task_projects_assignee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_projects
    ADD CONSTRAINT task_projects_assignee_id_fkey FOREIGN KEY (assignee_id) REFERENCES public.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: task_projects task_projects_assigner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_projects
    ADD CONSTRAINT task_projects_assigner_id_fkey FOREIGN KEY (assigner_id) REFERENCES public.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: task_projects task_projects_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_projects
    ADD CONSTRAINT task_projects_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.team_members(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: task_projects task_projects_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_projects
    ADD CONSTRAINT task_projects_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tasks tasks_assignee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_assignee_id_fkey FOREIGN KEY (assignee_id) REFERENCES public.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tasks tasks_assigner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_assigner_id_fkey FOREIGN KEY (assigner_id) REFERENCES public.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tasks tasks_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.team_members(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tasks tasks_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.tasks(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tasks tasks_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.task_projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tasks tasks_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_invitations team_invitations_inviter_member_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_invitations
    ADD CONSTRAINT team_invitations_inviter_member_id_fkey FOREIGN KEY (inviter_member_id) REFERENCES public.team_members(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_invitations team_invitations_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_invitations
    ADD CONSTRAINT team_invitations_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_members team_members_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_members team_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tokens tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tokens
    ADD CONSTRAINT tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_accounts user_accounts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_accounts
    ADD CONSTRAINT user_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_permissions user_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_permissions
    ADD CONSTRAINT user_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_permissions user_permissions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_permissions
    ADD CONSTRAINT user_permissions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20250321020346'),
    ('20250321105039'),
    ('20250321105959'),
    ('20250321112511'),
    ('20250321181227'),
    ('20250331070805'),
    ('20250331070806'),
    ('20250404060015'),
    ('20250410185852'),
    ('20250413052327'),
    ('20250414165202'),
    ('20250419024345'),
    ('20250505071914'),
    ('20250523035749');
