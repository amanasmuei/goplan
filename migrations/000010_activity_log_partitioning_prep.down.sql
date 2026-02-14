-- Remove activity log partitioning preparation objects

DROP FUNCTION IF EXISTS purge_archived_activity_logs(INTERVAL);
DROP FUNCTION IF EXISTS archive_old_activity_logs(INTERVAL);
DROP TABLE IF EXISTS activity_log_archive CASCADE;
DROP INDEX IF EXISTS idx_activity_log_created_at;
