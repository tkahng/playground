-- migrate:up
create or replace function not_empty(input text) returns boolean language plpgsql stable as $$ begin return (char_length(input) > 0);
end;
$$;
create type public.provider_types as enum ('oauth', 'credentials');
create type public.providers as enum (
    'google',
    'apple',
    'facebook',
    'github',
    'credentials'
);
create type public.token_types as enum (
    'access_token',
    'recovery_token',
    'invite_token',
    'reauthentication_token',
    'refresh_token',
    'verification_token',
    'password_reset_token',
    'state_token'
);
create table if not exists public.tokens (
    id uuid primary key default gen_random_uuid(),
    type public.token_types not null,
    user_id uuid references public.users on delete cascade on update cascade,
    -- type text not null,
    -- email or username
    otp varchar(255),
    identifier text not null,
    expires timestamptz not null,
    token text not null unique,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    -- metadata jsonb,
    constraint tokens_type_identifier_token_not_empty check (
        not_empty(identifier)
        and not_empty(token)
    )
);
CREATE TRIGGER handle_tokens_updated_at before
update on public.tokens for each row execute procedure moddatetime(updated_at);
-- -------------- USER ACCOUNTS TABLE START -----------------------------------------------------------------------
create table if not exists public.user_accounts (
    id uuid primary key default gen_random_uuid(),
    "user_id" uuid not null references public.users on delete cascade on update cascade,
    type provider_types not null,
    provider providers not null,
    /**
     * This value depends on the type of the provider being used to create the account.
     * - oauth/oidc: The OAuth account's id, returned from the `profile()` callback.
     * - email: The user's email address.
     * - credentials: `id` returned from the `authorize()` callback
     */
    "provider_account_id" varchar(255) not null,
    password text,
    refresh_token text,
    access_token text,
    expires_at bigint,
    id_token text,
    scope text,
    session_state text,
    token_type text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    -- compound unique constraint on user_id and provider
    -- constraint user_accounts_type_identifier_token_not_empty check (
    --     char_length(type) > 0
    --     and char_length(identifier) > 0
    --     and char_length(token) > 0
    -- ),
    -- constraint user_accounts_user_id_type_provider_account_id_not_empty check ("user_id", type, provider, "provider_account_id"),
    constraint user_accounts_provider_provider_account_id_unique unique (provider, "provider_account_id"),
    constraint user_accounts_user_id_provider_unique unique ("user_id", provider)
);
CREATE TRIGGER handle_user_accounts_updated_at before
update on public.user_accounts for each row execute procedure moddatetime(updated_at);
-- -------------- USER SESSIONS TABLE START -----------------------------------------------------------------------
create table if not exists public.user_sessions (
    id uuid primary key default gen_random_uuid(),
    "user_id" uuid not null references public.users on delete cascade on update cascade,
    expires timestamptz not null,
    "session_token" varchar(255) not null unique,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint user_sessions_token_not_empty check (not_empty("session_token"))
);
CREATE TRIGGER handle_user_sessions_updated_at before
update on public.user_sessions for each row execute procedure moddatetime(updated_at);
-- migrate:down
drop trigger if exists handle_user_sessions_updated_at on public.user_sessions;
alter table public.user_sessions drop constraint if exists user_sessions_token_not_empty;
drop table if exists public.user_sessions;
drop trigger if exists handle_user_accounts_updated_at on public.user_accounts;
alter table public.user_accounts drop constraint if exists user_accounts_user_id_provider_unique;
drop table if exists public.user_accounts;
drop trigger if exists handle_tokens_updated_at on public.tokens;
alter table public.tokens drop constraint if exists tokens_type_identifier_token_not_empty;
drop table if exists public.tokens;
drop type if exists public.provider_types;
drop type if exists public.providers;
drop type if exists public.token_types;
drop function if exists not_empty(text);