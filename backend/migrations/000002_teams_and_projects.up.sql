-- 000002_teams_and_projects.up.sql
-- Add teams, team members, and project-team associations

CREATE TYPE team_role AS ENUM ('owner', 'manager', 'member', 'viewer');

-- Teams table
CREATE TABLE teams (
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
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role team_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_per_team UNIQUE(team_id, user_id)
);

-- Project-Team association table
CREATE TABLE project_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_project_team UNIQUE(project_id, team_id)
);

-- Team indexes
CREATE INDEX idx_teams_org ON teams(organization_id);
CREATE INDEX idx_teams_created_by ON teams(created_by);
CREATE INDEX idx_teams_name ON teams(name);

-- Team members indexes
CREATE INDEX idx_team_members_team ON team_members(team_id);
CREATE INDEX idx_team_members_user ON team_members(user_id);
CREATE INDEX idx_team_members_role ON team_members(role);

-- Project-team indexes
CREATE INDEX idx_project_teams_project ON project_teams(project_id);
CREATE INDEX idx_project_teams_team ON project_teams(team_id);
CREATE INDEX idx_project_teams_team_project ON project_teams(team_id, project_id);
