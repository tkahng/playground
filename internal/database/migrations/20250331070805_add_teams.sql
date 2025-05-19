-- migrate:up
-- teams table  ----------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teams (
    id uuid primary key default gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    stripe_customer_id TEXT UNIQUE,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
create trigger handle_teams_updated_at before
update on public.teams for each row execute procedure set_current_timestamp_updated_at();
-- team member roles enum ------------------------------------------------
CREATE TYPE team_member_role AS ENUM ('owner', 'member', 'guest');
-- team members table ------------------------------------------------------
CREATE TABLE IF NOT EXISTS team_members (
    id uuid primary key default gen_random_uuid(),
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id uuid REFERENCES users(id) ON DELETE
    SET NULL ON UPDATE CASCADE,
        active boolean NOT NULL DEFAULT true,
        role team_member_role NOT NULL,
        last_selected_at timestamptz not null default now(),
        created_at timestamptz not null default now(),
        updated_at timestamptz not null default now(),
        constraint team_members_user_id_team_id unique (user_id, team_id)
);
CREATE TRIGGER handle_team_members_updated_at BEFORE
UPDATE ON public.team_members FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
-- team invitation table ------------------------------------------------------
create type team_invitation_status as enum ('pending', 'accepted', 'declined', 'cancelled');
CREATE TABLE IF NOT EXISTS team_invitations (
    id uuid primary key default gen_random_uuid(),
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE ON UPDATE CASCADE,
    invited_by uuid NOT NULL REFERENCES team_members(id) ON DELETE CASCADE ON UPDATE CASCADE,
    email text NOT NULL,
    role team_member_role NOT NULL,
    token text NOT NULL UNIQUE,
    "status" team_invitation_status DEFAULT 'pending' NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint team_invitations_email_team_id unique (email, team_id)
);
-- migrate:down
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