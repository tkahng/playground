-- migrate:up
-- Team roles remain the stable membership enum for now, but permissions become
-- data-driven. This bridge lets product code check capabilities such as
-- team.members.invite instead of hard-coding every authorization decision to
-- owner/member/guest.
--
-- These permissions are intentionally scoped to org.team_role_permissions
-- instead of auth.permissions. The auth.* RBAC tables describe global/user
-- permissions; team permissions are membership-role capabilities.
create table if not exists org.team_role_permissions (
    role org.team_member_role not null,
    permission_name text not null,
    created_at timestamptz not null default clock_timestamp(),
    primary key (role, permission_name)
);

-- migrate:down
drop table if exists org.team_role_permissions;
