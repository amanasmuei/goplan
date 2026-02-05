-- Drop comments and activity_log tables
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP TABLE IF EXISTS activity_log CASCADE;
DROP TABLE IF EXISTS comments CASCADE;
