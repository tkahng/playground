-- migrate:up
-- create enums
create type auth.provider_types as enum ('oauth', 'credentials');
create type auth.providers as enum (
    'google',
    'apple',
    'facebook',
    'github',
    'credentials'
);
create type auth.token_types as enum (
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
-- create tokens table
create table if not exists auth.tokens (
    id uuid primary key default uuidv7(),
    type auth.token_types not null,
    user_id uuid references auth.users on delete cascade on update cascade,
    otp varchar(255),
    identifier text not null,
    expires timestamptz not null,
    token text not null unique,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp(),
    -- metadata jsonb,
    constraint tokens_type_identifier_token_not_empty check (
        utility.not_empty(identifier)
        and utility.not_empty(token)
    )
);
CREATE TRIGGER handle_auth_tokens_updated_at before
update on auth.tokens for each row execute procedure utility.set_current_timestamp_updated_at();
-- create user accounts table
create table if not exists auth.user_accounts (
    id uuid primary key default uuidv7(),
    user_id uuid not null references auth.users on delete cascade on update cascade,
    type provider_types not null,
    provider providers not null,
    /**
     * This value depends on the type of the provider being used to create the account.
     * - oauth/oidc: The OAuth account's id, returned from the `profile()` callback.
     * - email: The user's email address.
     * - credentials: `id` returned from the `authorize()` callback
     */
    provider_account_id varchar(255) not null,
    password text,
    refresh_token text,
    access_token text,
    expires_at bigint,
    id_token text,
    scope text,
    session_state text,
    token_type text,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp(),
    constraint user_accounts_provider_provider_account_id_unique unique (provider, provider_account_id),
    constraint user_accounts_user_id_provider_unique unique (user_id, provider)
);
CREATE TRIGGER handle_auth_user_accounts_updated_at before
update on auth.user_accounts for each row execute procedure utility.set_current_timestamp_updated_at();
-- create user sessions
create table if not exists auth.user_sessions (
    id uuid primary key default uuidv7(),
    user_id uuid not null references auth.users on delete cascade on update cascade,
    expires timestamptz not null,
    session_token varchar(255) not null unique,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz not null default clock_timestamp(),
    constraint user_sessions_token_not_empty check (not_empty(session_token))
);
CREATE TRIGGER handle_auth_user_sessions_updated_at before
update on auth.user_sessions for each row execute procedure utility.set_current_timestamp_updated_at();
-- migrate:down
drop trigger if exists handle_auth_user_sessions_updated_at on auth.user_sessions;
alter table auth.user_sessions drop constraint if exists user_sessions_token_not_empty;
drop table if exists auth.user_sessions;
drop trigger if exists handle_auth_user_accounts_updated_at on auth.user_accounts;
alter table auth.user_accounts drop constraint if exists user_accounts_user_id_provider_unique;
drop table if exists auth.user_accounts;
drop trigger if exists handle_auth_tokens_updated_at on auth.tokens;
alter table auth.tokens drop constraint if exists tokens_type_identifier_token_not_empty;
drop table if exists auth.tokens;
drop type if exists auth.provider_types;
drop type if exists auth.providers;
drop type if exists auth.token_types;