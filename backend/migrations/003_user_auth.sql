-- Migration: Add password authentication to users
-- This migration adds password_hash column for user authentication

-- Add password_hash column to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Create index for faster email lookups during login
CREATE INDEX IF NOT EXISTS idx_users_email_password ON users(email) WHERE password_hash IS NOT NULL;

-- Set default password for existing demo users (password: "password123")
-- bcrypt hash of "password123" with cost 10
UPDATE users
SET password_hash = '$2a$10$rD2kYexsOQGq40aMzIMOge3fqwLPiAyMIyN/UncdpPnfTEx44tDAa'
WHERE password_hash IS NULL;

-- Make password_hash NOT NULL for new users (optional, uncomment if desired)
-- ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
