-- Concurrent index creation for large tables.
--
-- IMPORTANT: CREATE INDEX CONCURRENTLY cannot run inside a transaction block.
-- When using golang-migrate, you must run this migration outside a transaction.
-- With golang-migrate CLI, this happens automatically for files that do not use
-- BEGIN/COMMIT. Some migration tools require special configuration.
--
-- If you encounter "CREATE INDEX CONCURRENTLY cannot run inside a transaction block",
-- run these statements manually outside a transaction.

-- Concurrent version of the tasks created_at index for zero-downtime deployments
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_created_at_desc_concurrent ON tasks(created_at DESC);

-- Concurrent version of the plans created_at index for zero-downtime deployments
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_plans_created_at_desc_concurrent ON plans(created_at DESC);

-- Concurrent composite index on activity_log for workspace activity feeds
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_workspace_created_concurrent ON activity_log(workspace_id, created_at DESC);
