-- 000004_strategy_pivot.up.sql
-- Strategy pivot: drop old task management tables, add strategic planning tables

-- Enable extensions (idempotent)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

-- ============================================================================
-- New enum types
-- ============================================================================

CREATE TYPE subscription_tier AS ENUM ('free', 'pro', 'pro_plus');

CREATE TYPE plan_status AS ENUM ('draft', 'generating', 'complete', 'archived');

CREATE TYPE section_type AS ENUM (
    'executive_brief',
    'strategic_context',
    'recommended_approach',
    'phased_execution',
    'immediate_action'
);

CREATE TYPE plan_category AS ENUM (
    'business',
    'saas',
    'event',
    'nonprofit',
    'personal',
    'education',
    'real_estate',
    'generic'
);

CREATE TYPE generation_status AS ENUM ('pending', 'in_progress', 'completed', 'failed');

-- ============================================================================
-- Alter users table
-- ============================================================================

ALTER TABLE users ADD COLUMN subscription_tier subscription_tier NOT NULL DEFAULT 'free';
ALTER TABLE users ADD COLUMN stripe_customer_id TEXT;

-- ============================================================================
-- Subscriptions table
-- ============================================================================

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier subscription_tier NOT NULL DEFAULT 'free',
    max_plans INTEGER NOT NULL DEFAULT 3,
    max_regenerations_per_day INTEGER NOT NULL DEFAULT 5,
    can_export BOOLEAN NOT NULL DEFAULT FALSE,
    can_version_history BOOLEAN NOT NULL DEFAULT FALSE,
    can_refine BOOLEAN NOT NULL DEFAULT FALSE,
    stripe_subscription_id TEXT,
    period_start TIMESTAMP WITH TIME ZONE,
    period_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_subscription UNIQUE(user_id)
);

-- ============================================================================
-- Strategic plans table
-- ============================================================================

CREATE TABLE strategic_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    original_input TEXT NOT NULL,
    category plan_category NOT NULL DEFAULT 'generic',
    sub_category VARCHAR(255),
    complexity INTEGER,
    status plan_status NOT NULL DEFAULT 'draft',
    current_version INTEGER NOT NULL DEFAULT 1,
    content_embedding vector(384),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- Plan sections table
-- ============================================================================

CREATE TABLE plan_sections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id UUID NOT NULL REFERENCES strategic_plans(id) ON DELETE CASCADE,
    section_type section_type NOT NULL,
    section_order INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    content JSONB NOT NULL DEFAULT '{}',
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_plan_section UNIQUE(plan_id, section_type)
);

-- ============================================================================
-- Section versions table
-- ============================================================================

CREATE TABLE section_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    section_id UUID NOT NULL REFERENCES plan_sections(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES strategic_plans(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    content JSONB NOT NULL DEFAULT '{}',
    refinement_context TEXT,
    generated_by VARCHAR(255),
    token_usage JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_section_version UNIQUE(section_id, version)
);

-- ============================================================================
-- Plan versions table
-- ============================================================================

CREATE TABLE plan_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id UUID NOT NULL REFERENCES strategic_plans(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    snapshot JSONB NOT NULL DEFAULT '{}',
    change_summary TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_plan_version UNIQUE(plan_id, version)
);

-- ============================================================================
-- Generation log table
-- ============================================================================

CREATE TABLE generation_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id UUID REFERENCES strategic_plans(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    section_type section_type,
    status generation_status NOT NULL DEFAULT 'pending',
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    model VARCHAR(255),
    duration_ms INTEGER,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- Indexes
-- ============================================================================

-- Strategic plans indexes
CREATE INDEX idx_strategic_plans_user ON strategic_plans(user_id);
CREATE INDEX idx_strategic_plans_org ON strategic_plans(organization_id);
CREATE INDEX idx_strategic_plans_status ON strategic_plans(status);
CREATE INDEX idx_strategic_plans_category ON strategic_plans(category);
CREATE INDEX idx_strategic_plans_created_at_desc ON strategic_plans(created_at DESC);

-- HNSW index for pgvector similarity search on strategic plans
CREATE INDEX idx_strategic_plans_embedding_hnsw ON strategic_plans
USING hnsw (content_embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Plan sections indexes
CREATE INDEX idx_plan_sections_plan ON plan_sections(plan_id);
CREATE INDEX idx_plan_sections_type ON plan_sections(section_type);

-- Section versions indexes
CREATE INDEX idx_section_versions_section ON section_versions(section_id);
CREATE INDEX idx_section_versions_plan ON section_versions(plan_id);

-- Plan versions indexes
CREATE INDEX idx_plan_versions_plan ON plan_versions(plan_id);

-- Generation log indexes
CREATE INDEX idx_generation_log_user ON generation_log(user_id);
CREATE INDEX idx_generation_log_plan ON generation_log(plan_id);
CREATE INDEX idx_generation_log_created_at_desc ON generation_log(created_at DESC);

-- Subscriptions indexes
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_tier ON subscriptions(tier);

-- ============================================================================
-- Drop old task management tables (order matters for foreign keys)
-- ============================================================================

DROP TABLE IF EXISTS task_acknowledgments CASCADE;
DROP TABLE IF EXISTS task_reviews CASCADE;
DROP TABLE IF EXISTS task_blockers CASCADE;
DROP TABLE IF EXISTS task_justifications CASCADE;
DROP TABLE IF EXISTS task_links CASCADE;
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS project_teams CASCADE;
DROP TABLE IF EXISTS team_members CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS projects CASCADE;

-- Drop old enum types no longer needed
DROP TYPE IF EXISTS acknowledgment_action;
DROP TYPE IF EXISTS project_status;
DROP TYPE IF EXISTS blocker_type;
DROP TYPE IF EXISTS link_type;
DROP TYPE IF EXISTS task_status;
DROP TYPE IF EXISTS team_role;
