-- 000001_init.up.sql
-- Initial schema: extensions, enums, core tables, indexes

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

-- Create enum types
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

CREATE TYPE user_role AS ENUM ('admin', 'team_lead', 'member');

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

-- Organizations table
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'member',
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    password_hash TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

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

-- Tasks table (main entity)
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

-- Task links table (relationships between tasks)
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

-- Task justifications table (when no links exist)
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

-- Task reviews table (mandatory completion reviews)
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

-- Task acknowledgments table (records how user acknowledged predictions)
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

-- Indexes

-- HNSW index for pgvector similarity search (preferred for production)
CREATE INDEX idx_tasks_embedding_hnsw ON tasks
USING hnsw (description_embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Composite indexes for common queries
CREATE INDEX idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX idx_tasks_created_by_date ON tasks(created_by, created_at DESC);
CREATE INDEX idx_tasks_org_status ON tasks(organization_id, status);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_created_at_desc ON tasks(created_at DESC);

-- Link lookup indexes
CREATE INDEX idx_task_links_source ON task_links(source_task_id);
CREATE INDEX idx_task_links_target ON task_links(target_task_id);

-- Blocker indexes
CREATE INDEX idx_task_blockers_type_date ON task_blockers(blocker_type, created_at);
CREATE INDEX idx_task_blockers_task ON task_blockers(task_id);

-- User indexes
CREATE INDEX idx_users_org ON users(organization_id);
CREATE INDEX idx_users_email ON users(email);

-- Project indexes
CREATE INDEX idx_projects_org ON projects(organization_id);
CREATE INDEX idx_projects_status ON projects(status);

-- Acknowledgment index
CREATE INDEX idx_task_acknowledgments_task ON task_acknowledgments(task_id);
