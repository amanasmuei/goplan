-- 000003_user_auth.up.sql
-- Add password authentication to users

-- Create index for faster email lookups during login
CREATE INDEX idx_users_email_password ON users(email) WHERE password_hash IS NOT NULL;
