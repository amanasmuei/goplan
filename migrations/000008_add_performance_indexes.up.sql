-- Add performance indexes for common query patterns.
-- These indexes use CREATE INDEX IF NOT EXISTS for idempotency.
-- NOTE: For truly large tables, consider using CREATE INDEX CONCURRENTLY
-- (see migration 000009 for concurrent index creation).

-- =============================================================================
-- Tasks indexes
-- =============================================================================

-- Composite index on (plan_id, status) for filtering tasks within a plan by status.
-- The existing idx_tasks_status already covers (plan_id, status) so this is a no-op
-- if that index exists, but we keep the IF NOT EXISTS guard for safety.
CREATE INDEX IF NOT EXISTS idx_tasks_plan_status ON tasks(plan_id, status);

-- Index on assignee_id for looking up tasks assigned to a user.
-- Note: idx_tasks_assignee already exists from migration 000005, this is a safety net.
CREATE INDEX IF NOT EXISTS idx_tasks_assignee ON tasks(assignee_id);

-- Index on created_at DESC for sorting tasks by creation date (newest first).
CREATE INDEX IF NOT EXISTS idx_tasks_created_at_desc ON tasks(created_at DESC);

-- =============================================================================
-- Plans indexes
-- =============================================================================

-- Composite index on (workspace_id, status) for filtering plans within a workspace.
CREATE INDEX IF NOT EXISTS idx_plans_workspace_status ON plans(workspace_id, status);

-- Index on created_at DESC for sorting plans by creation date (newest first).
CREATE INDEX IF NOT EXISTS idx_plans_created_at_desc ON plans(created_at DESC);

-- =============================================================================
-- Activity log indexes
-- =============================================================================

-- Composite index on (workspace_id, created_at DESC) for time-ordered activity feeds.
CREATE INDEX IF NOT EXISTS idx_activity_workspace_created ON activity_log(workspace_id, created_at DESC);

-- =============================================================================
-- Comments indexes
-- =============================================================================

-- Composite index on (task_id, created_at) for ordered comment threads.
CREATE INDEX IF NOT EXISTS idx_comments_task_created ON comments(task_id, created_at);
