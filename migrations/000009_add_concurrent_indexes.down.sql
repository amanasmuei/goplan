-- Remove concurrent indexes added in 000009_add_concurrent_indexes.up.sql
DROP INDEX IF EXISTS idx_activity_workspace_created_concurrent;
DROP INDEX IF EXISTS idx_plans_created_at_desc_concurrent;
DROP INDEX IF EXISTS idx_tasks_created_at_desc_concurrent;
