-- Drop workspaces tables and related objects
DROP TRIGGER IF EXISTS update_workspaces_updated_at ON workspaces;
DROP TABLE IF EXISTS workspace_members CASCADE;
DROP TABLE IF EXISTS workspaces CASCADE;
