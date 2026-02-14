-- Remove performance indexes added in 000008_add_performance_indexes.up.sql

DROP INDEX IF EXISTS idx_comments_task_created;
DROP INDEX IF EXISTS idx_activity_workspace_created;
DROP INDEX IF EXISTS idx_plans_created_at_desc;
DROP INDEX IF EXISTS idx_plans_workspace_status;
DROP INDEX IF EXISTS idx_tasks_created_at_desc;
DROP INDEX IF EXISTS idx_tasks_plan_status;
