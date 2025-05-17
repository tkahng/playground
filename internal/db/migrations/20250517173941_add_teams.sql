-- migrate:up
-- teams table  ----------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teams (
    id uuid primary key default gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    stripe_customer_id TEXT UNIQUE,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
create trigger handle_teams_updated_at before
update on public.teams for each row execute procedure moddatetime(updated_at);
-- team member roles enum ------------------------------------------------
CREATE TYPE team_member_role AS ENUM ('admin', 'member', 'guest');
-- team members table ------------------------------------------------------
CREATE TABLE IF NOT EXISTS team_members (
    id uuid primary key default gen_random_uuid(),
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id uuid REFERENCES users(id) ON DELETE
    SET NULL ON UPDATE CASCADE,
        role team_member_role NOT NULL,
        created_at timestamptz not null default now(),
        updated_at timestamptz not null default now(),
        constraint team_members_user_id_team_id unique (user_id, team_id)
);
create trigger handle_team_members_updated_at before
update on public.team_members for each row execute procedure moddatetime(updated_at);
-- team invitation table ------------------------------------------------------
create type team_invitation_status as enum ('pending', 'accepted', 'declined');
CREATE TABLE IF NOT EXISTS team_invitations (
    id uuid primary key default gen_random_uuid(),
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE ON UPDATE CASCADE,
    invited_by uuid NOT NULL REFERENCES team_members(id) ON DELETE CASCADE ON UPDATE CASCADE,
    email text NOT NULL,
    role team_member_role NOT NULL,
    token text NOT NULL UNIQUE,
    "status" team_invitation_status DEFAULT 'pending' NOT NULL,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint team_invitations_email_team_id unique (email, team_id)
);
-- change stripe_subscription table refrences
ALTER TABLE public.stripe_subscriptions
ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE public.stripe_subscriptions
ADD COLUMN team_id uuid REFERENCES public.teams ON DELETE CASCADE ON UPDATE CASCADE;
-- migrate:down
-- subscriptions table  ----------------------------------------------------------------------
ALTER TABLE public.stripe_subscriptions DROP COLUMN team_id;
ALTER TABLE public.stripe_subscriptions
ALTER COLUMN user_id
SET NOT NULL;
-- team invitation table ------------------------------------------------------
drop table if exists public.team_invitations;
-- team invitation status enum ------------------------------------------------
drop type if exists team_invitation_status;
-- team members table ------------------------------------------------------
drop trigger if exists handle_team_members_updated_at on public.team_members;
drop table if exists public.team_members;
-- team member roles enum ------------------------------------------------
drop type if exists team_member_role;
-- teams table  ----------------------------------------------------------------------
drop trigger if exists handle_teams_updated_at on public.teams;
drop table if exists public.teams;