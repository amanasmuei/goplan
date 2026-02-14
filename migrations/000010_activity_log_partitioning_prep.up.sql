-- Activity log partitioning preparation.
--
-- For high-volume deployments, the activity_log table should be partitioned
-- by created_at using PostgreSQL range partitioning. This migration adds
-- supporting indexes and a retention policy helper function.
--
-- PARTITIONING GUIDE (for when you're ready to partition):
-- 1. Create a new partitioned table:
--    CREATE TABLE activity_log_partitioned (
--        LIKE activity_log INCLUDING ALL
--    ) PARTITION BY RANGE (created_at);
--
-- 2. Create monthly partitions:
--    CREATE TABLE activity_log_y2026m01 PARTITION OF activity_log_partitioned
--        FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
--    CREATE TABLE activity_log_y2026m02 PARTITION OF activity_log_partitioned
--        FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
--    -- ... repeat for each month
--
-- 3. Migrate data:
--    INSERT INTO activity_log_partitioned SELECT * FROM activity_log;
--
-- 4. Swap tables:
--    ALTER TABLE activity_log RENAME TO activity_log_old;
--    ALTER TABLE activity_log_partitioned RENAME TO activity_log;
--
-- 5. Drop old table after verification:
--    DROP TABLE activity_log_old;

-- =============================================================================
-- Index to support future range partitioning on created_at
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_activity_log_created_at ON activity_log(created_at);

-- =============================================================================
-- Retention policy helper function
-- =============================================================================
-- Archives and deletes activity_log records older than the specified interval.
-- Usage: SELECT archive_old_activity_logs('90 days');
-- This moves old records to an archive table, then deletes them from the main table.

-- Create the archive table if it doesn't exist
CREATE TABLE IF NOT EXISTS activity_log_archive (
    LIKE activity_log INCLUDING ALL
);

-- Create or replace the retention policy function
CREATE OR REPLACE FUNCTION archive_old_activity_logs(retention_interval INTERVAL)
RETURNS INTEGER AS $$
DECLARE
    archived_count INTEGER;
BEGIN
    -- Move old records to archive table
    WITH moved AS (
        DELETE FROM activity_log
        WHERE created_at < NOW() - retention_interval
        RETURNING *
    )
    INSERT INTO activity_log_archive
    SELECT * FROM moved;

    GET DIAGNOSTICS archived_count = ROW_COUNT;

    RAISE NOTICE 'Archived % activity log records older than %', archived_count, retention_interval;
    RETURN archived_count;
END;
$$ LANGUAGE plpgsql;

-- Helper function to delete archived records older than a given interval
-- Usage: SELECT purge_archived_activity_logs('1 year');
CREATE OR REPLACE FUNCTION purge_archived_activity_logs(retention_interval INTERVAL)
RETURNS INTEGER AS $$
DECLARE
    purged_count INTEGER;
BEGIN
    DELETE FROM activity_log_archive
    WHERE created_at < NOW() - retention_interval;

    GET DIAGNOSTICS purged_count = ROW_COUNT;

    RAISE NOTICE 'Purged % archived activity log records older than %', purged_count, retention_interval;
    RETURN purged_count;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- COMMENT: Recommended cron job for retention (run via pg_cron or external scheduler):
--
--   -- Archive records older than 90 days (run daily):
--   SELECT archive_old_activity_logs('90 days');
--
--   -- Purge archived records older than 1 year (run monthly):
--   SELECT purge_archived_activity_logs('1 year');
-- =============================================================================
