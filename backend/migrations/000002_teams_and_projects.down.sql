-- 000002_teams_and_projects.down.sql

DROP TABLE IF EXISTS project_teams CASCADE;
DROP TABLE IF EXISTS team_members CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TYPE IF EXISTS team_role;
