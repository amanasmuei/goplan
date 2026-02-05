# Domain Types Contract

> This document defines the core domain entities. All teams (Backend, Frontend, Database) MUST follow these definitions.

## Entity Naming Conventions

- **Backend (Go)**: PascalCase for types, camelCase for JSON tags
- **Frontend (TypeScript)**: PascalCase for types, camelCase for properties
- **Database (PostgreSQL)**: snake_case for tables and columns

---

## Core Entities

### 1. User

```typescript
// TypeScript
interface User {
  id: string;           // UUID
  email: string;
  name: string;
  avatarUrl: string | null;
  createdAt: string;    // ISO 8601
  updatedAt: string;    // ISO 8601
}
```

```go
// Go
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    AvatarURL *string   `json:"avatarUrl,omitempty"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

```sql
-- PostgreSQL
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

### 2. Workspace

```typescript
// TypeScript
interface Workspace {
  id: string;
  name: string;
  slug: string;
  ownerId: string;
  settings: WorkspaceSettings;
  createdAt: string;
  updatedAt: string;
}

interface WorkspaceSettings {
  defaultView: 'kanban' | 'list' | 'calendar' | 'gantt';
  aiEnabled: boolean;
}

interface WorkspaceMember {
  workspaceId: string;
  userId: string;
  role: 'owner' | 'admin' | 'member' | 'viewer';
  joinedAt: string;
}
```

```go
// Go
type Workspace struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Slug      string            `json:"slug"`
    OwnerID   string            `json:"ownerId"`
    Settings  WorkspaceSettings `json:"settings"`
    CreatedAt time.Time         `json:"createdAt"`
    UpdatedAt time.Time         `json:"updatedAt"`
}

type WorkspaceSettings struct {
    DefaultView string `json:"defaultView"`
    AIEnabled   bool   `json:"aiEnabled"`
}

type WorkspaceMember struct {
    WorkspaceID string    `json:"workspaceId"`
    UserID      string    `json:"userId"`
    Role        string    `json:"role"`
    JoinedAt    time.Time `json:"joinedAt"`
}
```

```sql
-- PostgreSQL
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    settings JSONB DEFAULT '{"defaultView": "kanban", "aiEnabled": true}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE workspace_members (
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);
```

---

### 3. Plan

```typescript
// TypeScript
interface Plan {
  id: string;
  workspaceId: string;
  name: string;
  description: string | null;
  domain: PlanDomain;
  status: PlanStatus;
  ownerId: string;
  startDate: string | null;  // ISO 8601 date
  endDate: string | null;
  customStatuses: StatusDefinition[];
  customFields: FieldDefinition[];
  tags: string[];
  createdAt: string;
  updatedAt: string;
}

type PlanDomain = 'software' | 'event' | 'ops' | 'collection' | 'generic';
type PlanStatus = 'draft' | 'active' | 'on_hold' | 'completed' | 'archived';

interface StatusDefinition {
  id: string;
  name: string;
  color: string;       // Hex color
  order: number;
  isDefault: boolean;
  isDone: boolean;     // Marks task as completed
}

interface FieldDefinition {
  id: string;
  name: string;
  type: 'text' | 'number' | 'date' | 'select' | 'multiselect' | 'checkbox';
  options?: string[];  // For select/multiselect
  required: boolean;
}
```

```go
// Go
type Plan struct {
    ID             string             `json:"id"`
    WorkspaceID    string             `json:"workspaceId"`
    Name           string             `json:"name"`
    Description    *string            `json:"description,omitempty"`
    Domain         string             `json:"domain"`
    Status         string             `json:"status"`
    OwnerID        string             `json:"ownerId"`
    StartDate      *string            `json:"startDate,omitempty"`
    EndDate        *string            `json:"endDate,omitempty"`
    CustomStatuses []StatusDefinition `json:"customStatuses"`
    CustomFields   []FieldDefinition  `json:"customFields"`
    Tags           []string           `json:"tags"`
    CreatedAt      time.Time          `json:"createdAt"`
    UpdatedAt      time.Time          `json:"updatedAt"`
}

type StatusDefinition struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Color     string `json:"color"`
    Order     int    `json:"order"`
    IsDefault bool   `json:"isDefault"`
    IsDone    bool   `json:"isDone"`
}

type FieldDefinition struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Type     string   `json:"type"`
    Options  []string `json:"options,omitempty"`
    Required bool     `json:"required"`
}
```

```sql
-- PostgreSQL
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

CREATE INDEX idx_plans_workspace ON plans(workspace_id);
CREATE INDEX idx_plans_owner ON plans(owner_id);
CREATE INDEX idx_plans_status ON plans(status);
```

---

### 4. Phase (Optional grouping within Plan)

```typescript
// TypeScript
interface Phase {
  id: string;
  planId: string;
  name: string;
  description: string | null;
  order: number;
  startDate: string | null;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
}
```

```go
// Go
type Phase struct {
    ID          string    `json:"id"`
    PlanID      string    `json:"planId"`
    Name        string    `json:"name"`
    Description *string   `json:"description,omitempty"`
    Order       int       `json:"order"`
    StartDate   *string   `json:"startDate,omitempty"`
    EndDate     *string   `json:"endDate,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

```sql
-- PostgreSQL
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

CREATE INDEX idx_phases_plan ON phases(plan_id);
```

---

### 5. Task

```typescript
// TypeScript
interface Task {
  id: string;
  planId: string;
  phaseId: string | null;
  parentId: string | null;     // For subtasks
  title: string;
  description: string | null;
  status: string;              // References plan's customStatuses
  priority: TaskPriority;
  assigneeId: string | null;
  dueDate: string | null;
  estimatedHours: number | null;
  customFieldValues: Record<string, unknown>;
  tags: string[];
  position: number;            // For ordering
  createdAt: string;
  updatedAt: string;

  // Computed/joined fields (not stored)
  subtasks?: Task[];
  dependencies?: TaskDependency[];
  comments?: Comment[];
}

type TaskPriority = 'low' | 'medium' | 'high' | 'critical';

interface TaskDependency {
  taskId: string;
  dependsOnId: string;
  type: 'blocks' | 'blocked_by';
}
```

```go
// Go
type Task struct {
    ID                string                 `json:"id"`
    PlanID            string                 `json:"planId"`
    PhaseID           *string                `json:"phaseId,omitempty"`
    ParentID          *string                `json:"parentId,omitempty"`
    Title             string                 `json:"title"`
    Description       *string                `json:"description,omitempty"`
    Status            string                 `json:"status"`
    Priority          string                 `json:"priority"`
    AssigneeID        *string                `json:"assigneeId,omitempty"`
    DueDate           *string                `json:"dueDate,omitempty"`
    EstimatedHours    *float64               `json:"estimatedHours,omitempty"`
    CustomFieldValues map[string]interface{} `json:"customFieldValues"`
    Tags              []string               `json:"tags"`
    Position          int                    `json:"position"`
    CreatedAt         time.Time              `json:"createdAt"`
    UpdatedAt         time.Time              `json:"updatedAt"`

    // Joined fields
    Subtasks     []Task           `json:"subtasks,omitempty"`
    Dependencies []TaskDependency `json:"dependencies,omitempty"`
    Comments     []Comment        `json:"comments,omitempty"`
}

type TaskDependency struct {
    TaskID      string `json:"taskId"`
    DependsOnID string `json:"dependsOnId"`
    Type        string `json:"type"`
}
```

```sql
-- PostgreSQL
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    phase_id UUID REFERENCES phases(id) ON DELETE SET NULL,
    parent_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(100) NOT NULL DEFAULT 'todo',
    priority VARCHAR(50) DEFAULT 'medium',
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date DATE,
    estimated_hours DECIMAL(6,2),
    custom_field_values JSONB DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    position INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE task_dependencies (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, depends_on_id),
    CHECK (task_id != depends_on_id)
);

CREATE INDEX idx_tasks_plan ON tasks(plan_id);
CREATE INDEX idx_tasks_phase ON tasks(phase_id);
CREATE INDEX idx_tasks_parent ON tasks(parent_id);
CREATE INDEX idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX idx_tasks_status ON tasks(plan_id, status);
CREATE INDEX idx_tasks_due_date ON tasks(due_date) WHERE due_date IS NOT NULL;

-- Full-text search
CREATE INDEX idx_tasks_search ON tasks USING gin(
    to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, ''))
);
```

---

### 6. Milestone

```typescript
// TypeScript
interface Milestone {
  id: string;
  planId: string;
  name: string;
  description: string | null;
  dueDate: string;
  status: 'pending' | 'reached' | 'missed';
  linkedTaskIds: string[];
  createdAt: string;
  updatedAt: string;
}
```

```go
// Go
type Milestone struct {
    ID            string    `json:"id"`
    PlanID        string    `json:"planId"`
    Name          string    `json:"name"`
    Description   *string   `json:"description,omitempty"`
    DueDate       string    `json:"dueDate"`
    Status        string    `json:"status"`
    LinkedTaskIDs []string  `json:"linkedTaskIds"`
    CreatedAt     time.Time `json:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt"`
}
```

```sql
-- PostgreSQL
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

CREATE INDEX idx_milestones_plan ON milestones(plan_id);
```

---

### 7. Comment

```typescript
// TypeScript
interface Comment {
  id: string;
  taskId: string;
  userId: string;
  content: string;
  mentions: string[];       // User IDs mentioned
  createdAt: string;
  updatedAt: string;

  // Joined
  user?: User;
}
```

```go
// Go
type Comment struct {
    ID        string    `json:"id"`
    TaskID    string    `json:"taskId"`
    UserID    string    `json:"userId"`
    Content   string    `json:"content"`
    Mentions  []string  `json:"mentions"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`

    User *User `json:"user,omitempty"`
}
```

```sql
-- PostgreSQL
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    mentions UUID[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_comments_task ON comments(task_id);
```

---

### 8. Activity Log

```typescript
// TypeScript
interface ActivityLog {
  id: string;
  workspaceId: string;
  planId: string | null;
  taskId: string | null;
  userId: string;
  action: ActivityAction;
  details: Record<string, unknown>;
  createdAt: string;
}

type ActivityAction =
  | 'plan.created' | 'plan.updated' | 'plan.archived'
  | 'task.created' | 'task.updated' | 'task.deleted' | 'task.moved'
  | 'task.assigned' | 'task.completed'
  | 'comment.added'
  | 'member.invited' | 'member.removed';
```

```go
// Go
type ActivityLog struct {
    ID          string                 `json:"id"`
    WorkspaceID string                 `json:"workspaceId"`
    PlanID      *string                `json:"planId,omitempty"`
    TaskID      *string                `json:"taskId,omitempty"`
    UserID      string                 `json:"userId"`
    Action      string                 `json:"action"`
    Details     map[string]interface{} `json:"details"`
    CreatedAt   time.Time              `json:"createdAt"`
}
```

```sql
-- PostgreSQL
CREATE TABLE activity_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    plan_id UUID REFERENCES plans(id) ON DELETE CASCADE,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_activity_workspace ON activity_log(workspace_id);
CREATE INDEX idx_activity_plan ON activity_log(plan_id);
CREATE INDEX idx_activity_created ON activity_log(created_at DESC);
```

---

## Enums / Constants

### Shared Constants (All Teams Must Use)

```typescript
// TypeScript - shared/constants.ts
export const PLAN_DOMAINS = ['software', 'event', 'ops', 'collection', 'generic'] as const;
export const PLAN_STATUSES = ['draft', 'active', 'on_hold', 'completed', 'archived'] as const;
export const TASK_PRIORITIES = ['low', 'medium', 'high', 'critical'] as const;
export const MEMBER_ROLES = ['owner', 'admin', 'member', 'viewer'] as const;
export const DEFAULT_VIEWS = ['kanban', 'list', 'calendar', 'gantt'] as const;
```

```go
// Go - domain/shared/constants.go
const (
    PlanDomainSoftware   = "software"
    PlanDomainEvent      = "event"
    PlanDomainOps        = "ops"
    PlanDomainCollection = "collection"
    PlanDomainGeneric    = "generic"
)

const (
    PlanStatusDraft     = "draft"
    PlanStatusActive    = "active"
    PlanStatusOnHold    = "on_hold"
    PlanStatusCompleted = "completed"
    PlanStatusArchived  = "archived"
)

const (
    TaskPriorityLow      = "low"
    TaskPriorityMedium   = "medium"
    TaskPriorityHigh     = "high"
    TaskPriorityCritical = "critical"
)

const (
    RoleOwner  = "owner"
    RoleAdmin  = "admin"
    RoleMember = "member"
    RoleViewer = "viewer"
)
```

---

## Validation Rules

| Field | Rule |
|-------|------|
| `email` | Valid email format, max 255 chars |
| `name` | 1-255 characters |
| `workspace.slug` | Lowercase alphanumeric + hyphens, 3-50 chars |
| `plan.name` | 1-255 characters |
| `task.title` | 1-500 characters |
| `priority` | One of: low, medium, high, critical |
| `status colors` | Valid hex color (#RRGGBB) |
| `UUID fields` | Valid UUID v4 format |
| `dates` | ISO 8601 format (YYYY-MM-DD) |
| `timestamps` | ISO 8601 with timezone |

---

## Default Values

| Entity | Field | Default |
|--------|-------|---------|
| Plan | domain | "generic" |
| Plan | status | "draft" |
| Plan | customStatuses | To Do, In Progress, Done |
| Task | status | "todo" (first status) |
| Task | priority | "medium" |
| Task | position | 0 |
| Workspace | settings.defaultView | "kanban" |
| Workspace | settings.aiEnabled | true |
| WorkspaceMember | role | "member" |
