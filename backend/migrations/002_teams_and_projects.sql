-- Migration: Add teams, team members, and project-team associations
-- This migration adds team management features and enhances project functionality

-- Team role enum
DO $$ BEGIN
    CREATE TYPE team_role AS ENUM ('owner', 'manager', 'member', 'viewer');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Teams table
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_team_name_per_org UNIQUE(organization_id, name)
);

-- Team members table (junction table with role)
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role team_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_per_team UNIQUE(team_id, user_id)
);

-- Project-Team association table
CREATE TABLE IF NOT EXISTS project_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_project_team UNIQUE(project_id, team_id)
);

-- Project status enum
DO $$ BEGIN
    CREATE TYPE project_status AS ENUM ('active', 'archived');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Add new columns to projects table
ALTER TABLE projects ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';

-- Team indexes
CREATE INDEX IF NOT EXISTS idx_teams_org ON teams(organization_id);
CREATE INDEX IF NOT EXISTS idx_teams_created_by ON teams(created_by);
CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name);

-- Team members indexes
CREATE INDEX IF NOT EXISTS idx_team_members_team ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_role ON team_members(role);

-- Project-team indexes
CREATE INDEX IF NOT EXISTS idx_project_teams_project ON project_teams(project_id);
CREATE INDEX IF NOT EXISTS idx_project_teams_team ON project_teams(team_id);

-- Project indexes for new columns
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_created_by ON projects(created_by);

-- Insert sample team data for development (linked to demo org)
INSERT INTO teams (id, name, description, organization_id, created_by) VALUES
    ('77777777-7777-7777-7777-777777777777', 'Engineering', 'Core engineering team', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222'),
    ('88888888-8888-8888-8888-888888888888', 'Product', 'Product management team', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222')
ON CONFLICT DO NOTHING;

-- Add team members
INSERT INTO team_members (team_id, user_id, role) VALUES
    ('77777777-7777-7777-7777-777777777777', '22222222-2222-2222-2222-222222222222', 'owner'),
    ('77777777-7777-7777-7777-777777777777', '33333333-3333-3333-3333-333333333333', 'manager'),
    ('77777777-7777-7777-7777-777777777777', '44444444-4444-4444-4444-444444444444', 'member'),
    ('88888888-8888-8888-8888-888888888888', '22222222-2222-2222-2222-222222222222', 'owner'),
    ('88888888-8888-8888-8888-888888888888', '33333333-3333-3333-3333-333333333333', 'member')
ON CONFLICT DO NOTHING;

-- Assign teams to projects
INSERT INTO project_teams (project_id, team_id) VALUES
    ('55555555-5555-5555-5555-555555555555', '77777777-7777-7777-7777-777777777777'),
    ('55555555-5555-5555-5555-555555555555', '88888888-8888-8888-8888-888888888888'),
    ('66666666-6666-6666-6666-666666666666', '77777777-7777-7777-7777-777777777777')
ON CONFLICT DO NOTHING;

-- Update existing projects to have a created_by if null
UPDATE projects SET created_by = '22222222-2222-2222-2222-222222222222' WHERE created_by IS NULL;
