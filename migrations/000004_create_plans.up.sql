-- Create plans table
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    domain VARCHAR(50) NOT NULL DEFAULT 'generic',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    owner_id UUID NOT NULL REFERENCES users(id),
    start_date DATE,
    end_date DATE,
    custom_statuses JSONB DEFAULT '[
        {"id": "todo", "name": "To Do", "color": "#6b7280", "order": 0, "isDefault": true, "isDone": false},
        {"id": "in_progress", "name": "In Progress", "color": "#3b82f6", "order": 1, "isDefault": false, "isDone": false},
        {"id": "done", "name": "Done", "color": "#10b981", "order": 2, "isDefault": false, "isDone": true}
    ]',
    custom_fields JSONB DEFAULT '[]',
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create phases table (optional grouping within Plan)
CREATE TABLE phases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    "order" INTEGER DEFAULT 0,
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create milestones table
CREATE TABLE milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    due_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    linked_task_ids UUID[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for plans
CREATE INDEX idx_plans_workspace ON plans(workspace_id);
CREATE INDEX idx_plans_owner ON plans(owner_id);
CREATE INDEX idx_plans_status ON plans(status);

-- Create indexes for phases
CREATE INDEX idx_phases_plan ON phases(plan_id);

-- Create indexes for milestones
CREATE INDEX idx_milestones_plan ON milestones(plan_id);

-- Add updated_at triggers
CREATE TRIGGER update_plans_updated_at
    BEFORE UPDATE ON plans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_phases_updated_at
    BEFORE UPDATE ON phases
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_milestones_updated_at
    BEFORE UPDATE ON milestones
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
