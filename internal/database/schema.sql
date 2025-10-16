\restrict W88AtFBc1d0o0wGoc0gDtlZwIeB8Exj6GnhcwCMVUFZCuoj8TlnMhbyHrq8pcVw

-- Dumped from database version 18.0 (Debian 18.0-1.pgdg13+3)
-- Dumped by pg_dump version 18.0

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
-- Name: app; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA app;


--
-- Name: auth; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA auth;


--
-- Name: billing; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA billing;


--
-- Name: messaging; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA messaging;


--
-- Name: org; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA org;


--
-- Name: sayhello; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA sayhello;


--
-- Name: storage; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA storage;


--
-- Name: task; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA task;


--
-- Name: utility; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA utility;


--
-- Name: job_status; Type: TYPE; Schema: app; Owner: -
--

CREATE TYPE app.job_status AS ENUM (
    'pending',
    'processing',
    'done',
    'failed'
);


--
-- Name: provider_types; Type: TYPE; Schema: auth; Owner: -
--

CREATE TYPE auth.provider_types AS ENUM (
    'oauth',
    'credentials'
);


--
-- Name: providers; Type: TYPE; Schema: auth; Owner: -
--

CREATE TYPE auth.providers AS ENUM (
    'google',
    'apple',
    'facebook',
    'github',
    'credentials'
);


--
-- Name: token_types; Type: TYPE; Schema: auth; Owner: -
--

CREATE TYPE auth.token_types AS ENUM (
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
-- Name: stripe_customer_type; Type: TYPE; Schema: billing; Owner: -
--

CREATE TYPE billing.stripe_customer_type AS ENUM (
    'user',
    'team'
);


--
-- Name: stripe_pricing_plan_interval; Type: TYPE; Schema: billing; Owner: -
--

CREATE TYPE billing.stripe_pricing_plan_interval AS ENUM (
    'day',
    'week',
    'month',
    'year'
);


--
-- Name: stripe_pricing_type; Type: TYPE; Schema: billing; Owner: -
--

CREATE TYPE billing.stripe_pricing_type AS ENUM (
    'one_time',
    'recurring'
);


--
-- Name: stripe_subscription_status; Type: TYPE; Schema: billing; Owner: -
--

CREATE TYPE billing.stripe_subscription_status AS ENUM (
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
-- Name: team_invitation_status; Type: TYPE; Schema: org; Owner: -
--

CREATE TYPE org.team_invitation_status AS ENUM (
    'pending',
    'accepted',
    'declined',
    'cancelled'
);


--
-- Name: team_member_role; Type: TYPE; Schema: org; Owner: -
--

CREATE TYPE org.team_member_role AS ENUM (
    'owner',
    'member',
    'guest'
);


--
-- Name: task_project_status; Type: TYPE; Schema: task; Owner: -
--

CREATE TYPE task.task_project_status AS ENUM (
    'todo',
    'in_progress',
    'done'
);


--
-- Name: task_status; Type: TYPE; Schema: task; Owner: -
--

CREATE TYPE task.task_status AS ENUM (
    'todo',
    'in_progress',
    'done'
);


--
-- Name: not_empty(text); Type: FUNCTION; Schema: utility; Owner: -
--

CREATE FUNCTION utility.not_empty(input text) RETURNS boolean
    LANGUAGE plpgsql STABLE
    AS $$ begin return (char_length(input) > 0);
end;
$$;


--
-- Name: set_current_timestamp_updated_at(); Type: FUNCTION; Schema: utility; Owner: -
--

CREATE FUNCTION utility.set_current_timestamp_updated_at() RETURNS trigger
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
-- Name: ai_usages; Type: TABLE; Schema: app; Owner: -
--

CREATE TABLE app.ai_usages (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid NOT NULL,
    prompt_tokens bigint NOT NULL,
    completion_tokens bigint NOT NULL,
    total_tokens bigint NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: audit_logs; Type: TABLE; Schema: app; Owner: -
--

CREATE TABLE app.audit_logs (
    id uuid DEFAULT uuidv7() NOT NULL,
    level integer DEFAULT 0 NOT NULL,
    source text,
    message text NOT NULL,
    data jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: jobs; Type: TABLE; Schema: app; Owner: -
--

CREATE TABLE app.jobs (
    id uuid DEFAULT uuidv7() NOT NULL,
    kind text NOT NULL,
    unique_key text,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    status app.job_status DEFAULT 'pending'::app.job_status NOT NULL,
    run_after timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 3 NOT NULL,
    last_error text,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: params; Type: TABLE; Schema: app; Owner: -
--

CREATE TABLE app.params (
    id uuid DEFAULT uuidv7() NOT NULL,
    name text NOT NULL,
    value jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: permissions; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.permissions (
    id uuid DEFAULT uuidv7() NOT NULL,
    name character varying(150) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: role_permissions; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.role_permissions (
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.roles (
    id uuid DEFAULT uuidv7() NOT NULL,
    name character varying(150) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: tokens; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.tokens (
    id uuid DEFAULT uuidv7() NOT NULL,
    type auth.token_types NOT NULL,
    user_id uuid,
    otp character varying(255),
    identifier text NOT NULL,
    expires timestamp with time zone NOT NULL,
    token text NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    CONSTRAINT tokens_type_identifier_token_not_empty CHECK ((utility.not_empty(identifier) AND utility.not_empty(token)))
);


--
-- Name: user_accounts; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.user_accounts (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid NOT NULL,
    type auth.provider_types NOT NULL,
    provider auth.providers NOT NULL,
    provider_account_id character varying(255) NOT NULL,
    password text,
    refresh_token text,
    access_token text,
    expires_at bigint,
    id_token text,
    scope text,
    session_state text,
    token_type text,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: user_permissions; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.user_permissions (
    user_id uuid NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: user_roles; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.user_roles (
    user_id uuid NOT NULL,
    role_id uuid NOT NULL
);


--
-- Name: user_sessions; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.user_sessions (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid NOT NULL,
    expires timestamp with time zone NOT NULL,
    session_token character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    CONSTRAINT user_sessions_token_not_empty CHECK (utility.not_empty((session_token)::text))
);


--
-- Name: users; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.users (
    id uuid DEFAULT uuidv7() NOT NULL,
    email character varying NOT NULL,
    email_verified_at timestamp with time zone,
    name character varying,
    image text,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: product_permissions; Type: TABLE; Schema: billing; Owner: -
--

CREATE TABLE billing.product_permissions (
    product_id text NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: product_roles; Type: TABLE; Schema: billing; Owner: -
--

CREATE TABLE billing.product_roles (
    product_id text NOT NULL,
    role_id uuid NOT NULL
);


--
-- Name: stripe_customers; Type: TABLE; Schema: billing; Owner: -
--

CREATE TABLE billing.stripe_customers (
    id text NOT NULL,
    email text NOT NULL,
    name text,
    customer_type billing.stripe_customer_type NOT NULL,
    user_id uuid,
    team_id uuid,
    billing_address jsonb,
    payment_method jsonb,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    CONSTRAINT stripe_customers_only_one_reference_check CHECK ((((user_id IS NOT NULL) AND (customer_type = 'user'::billing.stripe_customer_type) AND (team_id IS NULL)) OR ((user_id IS NULL) AND (team_id IS NOT NULL) AND (customer_type = 'team'::billing.stripe_customer_type))))
);


--
-- Name: stripe_prices; Type: TABLE; Schema: billing; Owner: -
--

CREATE TABLE billing.stripe_prices (
    id text NOT NULL,
    product_id text NOT NULL,
    lookup_key text,
    active boolean DEFAULT false NOT NULL,
    unit_amount bigint,
    currency text NOT NULL,
    type billing.stripe_pricing_type NOT NULL,
    "interval" billing.stripe_pricing_plan_interval,
    interval_count bigint,
    trial_period_days bigint,
    metadata jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    CONSTRAINT stripe_prices_currency_check CHECK ((char_length(currency) = 3))
);


--
-- Name: stripe_products; Type: TABLE; Schema: billing; Owner: -
--

CREATE TABLE billing.stripe_products (
    id text NOT NULL,
    active boolean DEFAULT false NOT NULL,
    name text NOT NULL,
    description text,
    image text,
    metadata jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: stripe_subscriptions; Type: TABLE; Schema: billing; Owner: -
--

CREATE TABLE billing.stripe_subscriptions (
    id text NOT NULL,
    stripe_customer_id text NOT NULL,
    status billing.stripe_subscription_status NOT NULL,
    metadata jsonb NOT NULL,
    item_id text NOT NULL,
    price_id text NOT NULL,
    quantity bigint NOT NULL,
    cancel_at_period_end boolean DEFAULT false NOT NULL,
    created timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    current_period_start timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    current_period_end timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    ended_at timestamp with time zone,
    cancel_at timestamp with time zone,
    canceled_at timestamp with time zone,
    trial_start timestamp with time zone,
    trial_end timestamp with time zone,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: stripe_webhook_events; Type: TABLE; Schema: billing; Owner: -
--

CREATE TABLE billing.stripe_webhook_events (
    id text NOT NULL,
    type text NOT NULL,
    object_type text NOT NULL,
    object_stripe_id text NOT NULL,
    event_creation_date timestamp with time zone NOT NULL,
    request_id text,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: notifications; Type: TABLE; Schema: messaging; Owner: -
--

CREATE TABLE messaging.notifications (
    id uuid DEFAULT uuidv7() NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    read_at timestamp with time zone,
    channel text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    user_id uuid,
    team_member_id uuid,
    team_id uuid,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    type text NOT NULL
);


--
-- Name: team_invitations; Type: TABLE; Schema: org; Owner: -
--

CREATE TABLE org.team_invitations (
    id uuid DEFAULT uuidv7() NOT NULL,
    team_id uuid NOT NULL,
    inviter_member_id uuid NOT NULL,
    email text NOT NULL,
    role org.team_member_role NOT NULL,
    token text NOT NULL,
    status org.team_invitation_status DEFAULT 'pending'::org.team_invitation_status NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: team_members; Type: TABLE; Schema: org; Owner: -
--

CREATE TABLE org.team_members (
    id uuid DEFAULT uuidv7() NOT NULL,
    team_id uuid NOT NULL,
    user_id uuid,
    active boolean DEFAULT true NOT NULL,
    role org.team_member_role NOT NULL,
    has_billing_access boolean DEFAULT false NOT NULL,
    last_selected_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: teams; Type: TABLE; Schema: org; Owner: -
--

CREATE TABLE org.teams (
    id uuid DEFAULT uuidv7() NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: user_reactions; Type: TABLE; Schema: sayhello; Owner: -
--

CREATE TABLE sayhello.user_reactions (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid,
    type text NOT NULL,
    reaction text,
    ip_address text,
    country text,
    city text,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: media; Type: TABLE; Schema: storage; Owner: -
--

CREATE TABLE storage.media (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid,
    disk character varying(32) NOT NULL,
    directory character varying(255) NOT NULL,
    filename character varying(255) NOT NULL,
    original_filename character varying(255) NOT NULL,
    extension character varying(32) NOT NULL,
    mime_type character varying(128) NOT NULL,
    size bigint NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: task_projects; Type: TABLE; Schema: task; Owner: -
--

CREATE TABLE task.task_projects (
    id uuid DEFAULT uuidv7() NOT NULL,
    team_id uuid NOT NULL,
    created_by_member_id uuid,
    name text NOT NULL,
    description text,
    status task.task_project_status DEFAULT 'todo'::task.task_project_status NOT NULL,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    assignee_id uuid,
    reporter_id uuid,
    rank double precision DEFAULT 0.0 NOT NULL,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: tasks; Type: TABLE; Schema: task; Owner: -
--

CREATE TABLE task.tasks (
    id uuid DEFAULT uuidv7() NOT NULL,
    team_id uuid NOT NULL,
    created_by_member_id uuid,
    project_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    status task.task_status DEFAULT 'todo'::task.task_status NOT NULL,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    assignee_id uuid,
    reporter_id uuid,
    rank double precision DEFAULT 0.0 NOT NULL,
    parent_id uuid,
    created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL,
    updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);


--
-- Name: ai_usages ai_usages_pkey; Type: CONSTRAINT; Schema: app; Owner: -
--

ALTER TABLE ONLY app.ai_usages
    ADD CONSTRAINT ai_usages_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: app; Owner: -
--

ALTER TABLE ONLY app.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: app; Owner: -
--

ALTER TABLE ONLY app.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- Name: params params_name_key; Type: CONSTRAINT; Schema: app; Owner: -
--

ALTER TABLE ONLY app.params
    ADD CONSTRAINT params_name_key UNIQUE (name);


--
-- Name: params params_pkey; Type: CONSTRAINT; Schema: app; Owner: -
--

ALTER TABLE ONLY app.params
    ADD CONSTRAINT params_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_name_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.permissions
    ADD CONSTRAINT permissions_name_key UNIQUE (name);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_name_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: tokens tokens_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.tokens
    ADD CONSTRAINT tokens_pkey PRIMARY KEY (id);


--
-- Name: tokens tokens_token_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.tokens
    ADD CONSTRAINT tokens_token_key UNIQUE (token);


--
-- Name: user_accounts user_accounts_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_accounts
    ADD CONSTRAINT user_accounts_pkey PRIMARY KEY (id);


--
-- Name: user_accounts user_accounts_provider_provider_account_id_unique; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_accounts
    ADD CONSTRAINT user_accounts_provider_provider_account_id_unique UNIQUE (provider, provider_account_id);


--
-- Name: user_accounts user_accounts_user_id_provider_unique; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_accounts
    ADD CONSTRAINT user_accounts_user_id_provider_unique UNIQUE (user_id, provider);


--
-- Name: user_permissions user_permissions_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_permissions
    ADD CONSTRAINT user_permissions_pkey PRIMARY KEY (user_id, permission_id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_session_token_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_sessions
    ADD CONSTRAINT user_sessions_session_token_key UNIQUE (session_token);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: product_permissions product_permissions_pkey; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.product_permissions
    ADD CONSTRAINT product_permissions_pkey PRIMARY KEY (product_id, permission_id);


--
-- Name: product_roles product_roles_pkey; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.product_roles
    ADD CONSTRAINT product_roles_pkey PRIMARY KEY (product_id, role_id);


--
-- Name: stripe_customers stripe_customers_pkey; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_customers
    ADD CONSTRAINT stripe_customers_pkey PRIMARY KEY (id);


--
-- Name: stripe_customers stripe_customers_team_id_key; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_customers
    ADD CONSTRAINT stripe_customers_team_id_key UNIQUE (team_id);


--
-- Name: stripe_customers stripe_customers_user_id_key; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_customers
    ADD CONSTRAINT stripe_customers_user_id_key UNIQUE (user_id);


--
-- Name: stripe_prices stripe_prices_pkey; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_prices
    ADD CONSTRAINT stripe_prices_pkey PRIMARY KEY (id);


--
-- Name: stripe_products stripe_products_pkey; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_products
    ADD CONSTRAINT stripe_products_pkey PRIMARY KEY (id);


--
-- Name: stripe_subscriptions stripe_subscriptions_pkey; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_subscriptions
    ADD CONSTRAINT stripe_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: stripe_webhook_events stripe_webhook_events_pkey; Type: CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_webhook_events
    ADD CONSTRAINT stripe_webhook_events_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: team_invitations team_invitations_email_team_id; Type: CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_invitations
    ADD CONSTRAINT team_invitations_email_team_id UNIQUE (email, team_id);


--
-- Name: team_invitations team_invitations_pkey; Type: CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_invitations
    ADD CONSTRAINT team_invitations_pkey PRIMARY KEY (id);


--
-- Name: team_invitations team_invitations_token_key; Type: CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_invitations
    ADD CONSTRAINT team_invitations_token_key UNIQUE (token);


--
-- Name: team_members team_members_pkey; Type: CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_members
    ADD CONSTRAINT team_members_pkey PRIMARY KEY (id);


--
-- Name: team_members team_members_user_id_team_id; Type: CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_members
    ADD CONSTRAINT team_members_user_id_team_id UNIQUE (user_id, team_id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: teams teams_slug_key; Type: CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.teams
    ADD CONSTRAINT teams_slug_key UNIQUE (slug);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: user_reactions user_reactions_pkey; Type: CONSTRAINT; Schema: sayhello; Owner: -
--

ALTER TABLE ONLY sayhello.user_reactions
    ADD CONSTRAINT user_reactions_pkey PRIMARY KEY (id);


--
-- Name: media media_disk_directory_filename_extension; Type: CONSTRAINT; Schema: storage; Owner: -
--

ALTER TABLE ONLY storage.media
    ADD CONSTRAINT media_disk_directory_filename_extension UNIQUE (disk, directory, filename, extension);


--
-- Name: media media_pkey; Type: CONSTRAINT; Schema: storage; Owner: -
--

ALTER TABLE ONLY storage.media
    ADD CONSTRAINT media_pkey PRIMARY KEY (id);


--
-- Name: task_projects task_projects_pkey; Type: CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.task_projects
    ADD CONSTRAINT task_projects_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: idx_logs_created_at; Type: INDEX; Schema: app; Owner: -
--

CREATE INDEX idx_logs_created_at ON app.audit_logs USING btree (created_at);


--
-- Name: idx_logs_data_gin; Type: INDEX; Schema: app; Owner: -
--

CREATE INDEX idx_logs_data_gin ON app.audit_logs USING gin (data);


--
-- Name: idx_logs_level; Type: INDEX; Schema: app; Owner: -
--

CREATE INDEX idx_logs_level ON app.audit_logs USING btree (level);


--
-- Name: idx_logs_source; Type: INDEX; Schema: app; Owner: -
--

CREATE INDEX idx_logs_source ON app.audit_logs USING btree (source);


--
-- Name: jobs_polling_idx; Type: INDEX; Schema: app; Owner: -
--

CREATE INDEX jobs_polling_idx ON app.jobs USING btree (status, run_after, attempts);


--
-- Name: uniq_jobs_active_key; Type: INDEX; Schema: app; Owner: -
--

CREATE UNIQUE INDEX uniq_jobs_active_key ON app.jobs USING btree (unique_key) WHERE (status = ANY (ARRAY['pending'::app.job_status, 'processing'::app.job_status]));


--
-- Name: params handle_app_params_updated_at; Type: TRIGGER; Schema: app; Owner: -
--

CREATE TRIGGER handle_app_params_updated_at BEFORE UPDATE ON app.params FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: permissions handle_auth_permissions_updated_at; Type: TRIGGER; Schema: auth; Owner: -
--

CREATE TRIGGER handle_auth_permissions_updated_at BEFORE UPDATE ON auth.permissions FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: roles handle_auth_roles_updated_at; Type: TRIGGER; Schema: auth; Owner: -
--

CREATE TRIGGER handle_auth_roles_updated_at BEFORE UPDATE ON auth.roles FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: tokens handle_auth_tokens_updated_at; Type: TRIGGER; Schema: auth; Owner: -
--

CREATE TRIGGER handle_auth_tokens_updated_at BEFORE UPDATE ON auth.tokens FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: user_accounts handle_auth_user_accounts_updated_at; Type: TRIGGER; Schema: auth; Owner: -
--

CREATE TRIGGER handle_auth_user_accounts_updated_at BEFORE UPDATE ON auth.user_accounts FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: user_sessions handle_auth_user_sessions_updated_at; Type: TRIGGER; Schema: auth; Owner: -
--

CREATE TRIGGER handle_auth_user_sessions_updated_at BEFORE UPDATE ON auth.user_sessions FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: users handle_auth_users_updated_at; Type: TRIGGER; Schema: auth; Owner: -
--

CREATE TRIGGER handle_auth_users_updated_at BEFORE UPDATE ON auth.users FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: stripe_products handle_messaging_notifications_updated_at; Type: TRIGGER; Schema: billing; Owner: -
--

CREATE TRIGGER handle_messaging_notifications_updated_at BEFORE UPDATE ON billing.stripe_products FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: stripe_customers handle_stripe_customers_updated_at; Type: TRIGGER; Schema: billing; Owner: -
--

CREATE TRIGGER handle_stripe_customers_updated_at BEFORE UPDATE ON billing.stripe_customers FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: stripe_prices handle_stripe_prices_updated_at; Type: TRIGGER; Schema: billing; Owner: -
--

CREATE TRIGGER handle_stripe_prices_updated_at BEFORE UPDATE ON billing.stripe_prices FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: stripe_products handle_stripe_products_updated_at; Type: TRIGGER; Schema: billing; Owner: -
--

CREATE TRIGGER handle_stripe_products_updated_at BEFORE UPDATE ON billing.stripe_products FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: stripe_subscriptions handle_stripe_subscriptions_updated_at; Type: TRIGGER; Schema: billing; Owner: -
--

CREATE TRIGGER handle_stripe_subscriptions_updated_at BEFORE UPDATE ON billing.stripe_subscriptions FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: stripe_webhook_events handle_stripe_webhook_events_updated_at; Type: TRIGGER; Schema: billing; Owner: -
--

CREATE TRIGGER handle_stripe_webhook_events_updated_at BEFORE UPDATE ON billing.stripe_webhook_events FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: team_invitations handle_team_invitations_updated_at; Type: TRIGGER; Schema: org; Owner: -
--

CREATE TRIGGER handle_team_invitations_updated_at BEFORE UPDATE ON org.team_invitations FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: team_members handle_team_members_updated_at; Type: TRIGGER; Schema: org; Owner: -
--

CREATE TRIGGER handle_team_members_updated_at BEFORE UPDATE ON org.team_members FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: teams handle_teams_updated_at; Type: TRIGGER; Schema: org; Owner: -
--

CREATE TRIGGER handle_teams_updated_at BEFORE UPDATE ON org.teams FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: user_reactions handle_user_reactions_updated_at; Type: TRIGGER; Schema: sayhello; Owner: -
--

CREATE TRIGGER handle_user_reactions_updated_at BEFORE UPDATE ON sayhello.user_reactions FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: media handle_media_updated_at; Type: TRIGGER; Schema: storage; Owner: -
--

CREATE TRIGGER handle_media_updated_at BEFORE UPDATE ON storage.media FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: task_projects handle_task_projects_updated_at; Type: TRIGGER; Schema: task; Owner: -
--

CREATE TRIGGER handle_task_projects_updated_at BEFORE UPDATE ON task.task_projects FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: tasks handle_tasks_updated_at; Type: TRIGGER; Schema: task; Owner: -
--

CREATE TRIGGER handle_tasks_updated_at BEFORE UPDATE ON task.tasks FOR EACH ROW EXECUTE FUNCTION utility.set_current_timestamp_updated_at();


--
-- Name: ai_usages ai_usages_user_id_fkey; Type: FK CONSTRAINT; Schema: app; Owner: -
--

ALTER TABLE ONLY app.ai_usages
    ADD CONSTRAINT ai_usages_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES auth.permissions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES auth.roles(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tokens tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.tokens
    ADD CONSTRAINT tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_accounts user_accounts_user_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_accounts
    ADD CONSTRAINT user_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_permissions user_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_permissions
    ADD CONSTRAINT user_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES auth.permissions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_permissions user_permissions_user_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_permissions
    ADD CONSTRAINT user_permissions_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES auth.roles(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product_permissions product_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.product_permissions
    ADD CONSTRAINT product_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES auth.permissions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product_permissions product_permissions_product_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.product_permissions
    ADD CONSTRAINT product_permissions_product_id_fkey FOREIGN KEY (product_id) REFERENCES billing.stripe_products(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product_roles product_roles_product_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.product_roles
    ADD CONSTRAINT product_roles_product_id_fkey FOREIGN KEY (product_id) REFERENCES billing.stripe_products(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product_roles product_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.product_roles
    ADD CONSTRAINT product_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES auth.roles(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stripe_customers stripe_customers_team_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_customers
    ADD CONSTRAINT stripe_customers_team_id_fkey FOREIGN KEY (team_id) REFERENCES org.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stripe_customers stripe_customers_user_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_customers
    ADD CONSTRAINT stripe_customers_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stripe_prices stripe_prices_product_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_prices
    ADD CONSTRAINT stripe_prices_product_id_fkey FOREIGN KEY (product_id) REFERENCES billing.stripe_products(id);


--
-- Name: stripe_subscriptions stripe_subscriptions_price_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_subscriptions
    ADD CONSTRAINT stripe_subscriptions_price_id_fkey FOREIGN KEY (price_id) REFERENCES billing.stripe_prices(id);


--
-- Name: stripe_subscriptions stripe_subscriptions_stripe_customer_id_fkey; Type: FK CONSTRAINT; Schema: billing; Owner: -
--

ALTER TABLE ONLY billing.stripe_subscriptions
    ADD CONSTRAINT stripe_subscriptions_stripe_customer_id_fkey FOREIGN KEY (stripe_customer_id) REFERENCES billing.stripe_customers(id);


--
-- Name: notifications fk_notifications_team; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT fk_notifications_team FOREIGN KEY (team_id) REFERENCES org.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notifications fk_notifications_team_member; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT fk_notifications_team_member FOREIGN KEY (team_member_id) REFERENCES org.team_members(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notifications fk_notifications_user; Type: FK CONSTRAINT; Schema: messaging; Owner: -
--

ALTER TABLE ONLY messaging.notifications
    ADD CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_invitations team_invitations_inviter_member_id_fkey; Type: FK CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_invitations
    ADD CONSTRAINT team_invitations_inviter_member_id_fkey FOREIGN KEY (inviter_member_id) REFERENCES org.team_members(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_invitations team_invitations_team_id_fkey; Type: FK CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_invitations
    ADD CONSTRAINT team_invitations_team_id_fkey FOREIGN KEY (team_id) REFERENCES org.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_members team_members_team_id_fkey; Type: FK CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_members
    ADD CONSTRAINT team_members_team_id_fkey FOREIGN KEY (team_id) REFERENCES org.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_members team_members_user_id_fkey; Type: FK CONSTRAINT; Schema: org; Owner: -
--

ALTER TABLE ONLY org.team_members
    ADD CONSTRAINT team_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: user_reactions user_reactions_user_id_fkey; Type: FK CONSTRAINT; Schema: sayhello; Owner: -
--

ALTER TABLE ONLY sayhello.user_reactions
    ADD CONSTRAINT user_reactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: media media_user_id_fkey; Type: FK CONSTRAINT; Schema: storage; Owner: -
--

ALTER TABLE ONLY storage.media
    ADD CONSTRAINT media_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: task_projects task_projects_assignee_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.task_projects
    ADD CONSTRAINT task_projects_assignee_id_fkey FOREIGN KEY (assignee_id) REFERENCES org.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: task_projects task_projects_created_by_member_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.task_projects
    ADD CONSTRAINT task_projects_created_by_member_id_fkey FOREIGN KEY (created_by_member_id) REFERENCES org.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: task_projects task_projects_reporter_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.task_projects
    ADD CONSTRAINT task_projects_reporter_id_fkey FOREIGN KEY (reporter_id) REFERENCES org.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: task_projects task_projects_team_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.task_projects
    ADD CONSTRAINT task_projects_team_id_fkey FOREIGN KEY (team_id) REFERENCES org.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tasks tasks_assignee_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.tasks
    ADD CONSTRAINT tasks_assignee_id_fkey FOREIGN KEY (assignee_id) REFERENCES org.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tasks tasks_created_by_member_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.tasks
    ADD CONSTRAINT tasks_created_by_member_id_fkey FOREIGN KEY (created_by_member_id) REFERENCES org.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tasks tasks_parent_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.tasks
    ADD CONSTRAINT tasks_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES task.tasks(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tasks tasks_project_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.tasks
    ADD CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES task.task_projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tasks tasks_reporter_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.tasks
    ADD CONSTRAINT tasks_reporter_id_fkey FOREIGN KEY (reporter_id) REFERENCES org.team_members(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tasks tasks_team_id_fkey; Type: FK CONSTRAINT; Schema: task; Owner: -
--

ALTER TABLE ONLY task.tasks
    ADD CONSTRAINT tasks_team_id_fkey FOREIGN KEY (team_id) REFERENCES org.teams(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict W88AtFBc1d0o0wGoc0gDtlZwIeB8Exj6GnhcwCMVUFZCuoj8TlnMhbyHrq8pcVw


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20250321020346'),
    ('20250321105038'),
    ('20250321105039'),
    ('20250321105959'),
    ('20250321112511'),
    ('20250321181226'),
    ('20250321181227'),
    ('20250331070804'),
    ('20250331070805'),
    ('20250331070807'),
    ('20250331070808'),
    ('20250404060014'),
    ('20250404060015'),
    ('20250410185851'),
    ('20250410185852'),
    ('20250413052326'),
    ('20250413052327'),
    ('20250414165202'),
    ('20250419024345'),
    ('20250505071914'),
    ('20250523035749'),
    ('20250717035204'),
    ('20250717035205');
