-- Create MCP (Model Context Protocol) audit table for AI interactions
CREATE TABLE mcp_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    session_id VARCHAR(255),
    request_type VARCHAR(100) NOT NULL,
    request_payload JSONB,
    response_payload JSONB,
    model VARCHAR(100),
    tokens_used INTEGER,
    latency_ms INTEGER,
    status VARCHAR(50) NOT NULL DEFAULT 'success',
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for mcp_audit_log
CREATE INDEX idx_mcp_audit_workspace ON mcp_audit_log(workspace_id);
CREATE INDEX idx_mcp_audit_user ON mcp_audit_log(user_id);
CREATE INDEX idx_mcp_audit_session ON mcp_audit_log(session_id);
CREATE INDEX idx_mcp_audit_created ON mcp_audit_log(created_at DESC);
CREATE INDEX idx_mcp_audit_status ON mcp_audit_log(status);

-- Create AI suggestions table for tracking AI-generated recommendations
CREATE TABLE ai_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    plan_id UUID REFERENCES plans(id) ON DELETE CASCADE,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    suggestion_type VARCHAR(100) NOT NULL,
    content JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    accepted_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for ai_suggestions
CREATE INDEX idx_ai_suggestions_workspace ON ai_suggestions(workspace_id);
CREATE INDEX idx_ai_suggestions_plan ON ai_suggestions(plan_id);
CREATE INDEX idx_ai_suggestions_task ON ai_suggestions(task_id);
CREATE INDEX idx_ai_suggestions_user ON ai_suggestions(user_id);
CREATE INDEX idx_ai_suggestions_status ON ai_suggestions(status);
