-- 000004_strategy_pivot.down.sql
-- Reverse strategy pivot: drop new tables, recreate old task management tables

-- ============================================================================
-- Drop new tables (order matters for foreign keys)
-- ============================================================================

DROP TABLE IF EXISTS generation_log CASCADE;
DROP TABLE IF EXISTS plan_versions CASCADE;
DROP TABLE IF EXISTS section_versions CASCADE;
DROP TABLE IF EXISTS plan_sections CASCADE;
DROP TABLE IF EXISTS strategic_plans CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;

-- ============================================================================
-- Remove columns added to users
-- ============================================================================

ALTER TABLE users DROP COLUMN IF EXISTS stripe_customer_id;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_tier;

-- ============================================================================
-- Drop new enum types
-- ============================================================================

DROP TYPE IF EXISTS generation_status;
DROP TYPE IF EXISTS plan_category;
DROP TYPE IF EXISTS section_type;
DROP TYPE IF EXISTS plan_status;
DROP TYPE IF EXISTS subscription_tier;

-- ============================================================================
-- Recreate old enum types
-- ============================================================================

CREATE TYPE task_status AS ENUM (
    'draft',
    'pending_acknowledgment',
    'acknowledged',
    'active',
    'blocked',
    'pending_review',
    'completed',
    'cancelled'
);

CREATE TYPE link_type AS ENUM ('similar', 'dependent', 'retry', 'related');

CREATE TYPE blocker_type AS ENUM (
    'approval',
    'external_team',
    'vendor',
    'technical',
    'resource',
    'requirements'
);

CREATE TYPE project_status AS ENUM ('active', 'archived');

CREATE TYPE acknowledgment_action AS ENUM ('accept', 'modify', 'disagree');

CREATE TYPE team_role AS ENUM ('owner', 'manager', 'member', 'viewer');

-- ============================================================================
-- Recreate old tables
-- ============================================================================

-- Projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status project_status NOT NULL DEFAULT 'active',
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    description_embedding vector(384),
    status task_status NOT NULL DEFAULT 'draft',
    created_by UUID NOT NULL REFERENCES users(id),
    assigned_to UUID REFERENCES users(id),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    estimated_days DECIMAL(10,2),
    predicted_days_low DECIMAL(10,2),
    predicted_days_high DECIMAL(10,2),
    prediction_confidence DECIMAL(3,2),
    planning_quality_score DECIMAL(5,2),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    actual_days DECIMAL(10,2),
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Task links table
CREATE TABLE task_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    target_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    link_type link_type NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_task_link UNIQUE (source_task_id, target_task_id)
);

-- Task justifications table
CREATE TABLE task_justifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL UNIQUE REFERENCES tasks(id) ON DELETE CASCADE,
    checked_same_project BOOLEAN NOT NULL DEFAULT FALSE,
    checked_same_stakeholders BOOLEAN NOT NULL DEFAULT FALSE,
    checked_same_dependencies BOOLEAN NOT NULL DEFAULT FALSE,
    justification_text TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Task blockers table
CREATE TABLE task_blockers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    blocker_type blocker_type NOT NULL,
    description TEXT NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    days_blocked DECIMAL(10,2),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Task reviews table
CREATE TABLE task_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL UNIQUE REFERENCES tasks(id) ON DELETE CASCADE,
    prediction_accuracy_rating INTEGER NOT NULL CHECK (prediction_accuracy_rating BETWEEN 1 AND 5),
    prediction_feedback TEXT,
    lessons_learned TEXT,
    would_approach_differently TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Task acknowledgments table
CREATE TABLE task_acknowledgments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    action acknowledgment_action NOT NULL,
    original_estimate DECIMAL(10,2),
    modified_estimate DECIMAL(10,2),
    predicted_low DECIMAL(10,2),
    predicted_high DECIMAL(10,2),
    disagreement_notes TEXT,
    acknowledged_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

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

-- Team members table
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role team_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_per_team UNIQUE(team_id, user_id)
);

-- Project-team association table
CREATE TABLE project_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_project_team UNIQUE(project_id, team_id)
);

-- ============================================================================
-- Recreate old indexes
-- ============================================================================

-- Task indexes
CREATE INDEX idx_tasks_embedding_hnsw ON tasks
USING hnsw (description_embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

CREATE INDEX idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX idx_tasks_created_by_date ON tasks(created_by, created_at DESC);
CREATE INDEX idx_tasks_org_status ON tasks(organization_id, status);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_created_at_desc ON tasks(created_at DESC);

-- Link indexes
CREATE INDEX idx_task_links_source ON task_links(source_task_id);
CREATE INDEX idx_task_links_target ON task_links(target_task_id);

-- Blocker indexes
CREATE INDEX idx_task_blockers_type_date ON task_blockers(blocker_type, created_at);
CREATE INDEX idx_task_blockers_task ON task_blockers(task_id);

-- Project indexes
CREATE INDEX idx_projects_org ON projects(organization_id);
CREATE INDEX idx_projects_status ON projects(status);

-- Acknowledgment index
CREATE INDEX idx_task_acknowledgments_task ON task_acknowledgments(task_id);

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
