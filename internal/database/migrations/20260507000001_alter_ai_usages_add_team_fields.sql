-- migrate:up
ALTER TABLE app.ai_usages
  ADD COLUMN team_member_id uuid REFERENCES org.team_members(id) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD COLUMN team_id        uuid REFERENCES org.teams(id)        ON DELETE SET NULL ON UPDATE CASCADE;

CREATE INDEX ai_usages_team_member_id_created_at_idx ON app.ai_usages(team_member_id, created_at)
  WHERE team_member_id IS NOT NULL;

CREATE INDEX ai_usages_team_id_created_at_idx ON app.ai_usages(team_id, created_at)
  WHERE team_id IS NOT NULL;

-- migrate:down
DROP INDEX IF EXISTS ai_usages_team_member_id_created_at_idx;
DROP INDEX IF EXISTS ai_usages_team_id_created_at_idx;
ALTER TABLE app.ai_usages
  DROP COLUMN IF EXISTS team_member_id,
  DROP COLUMN IF EXISTS team_id;
