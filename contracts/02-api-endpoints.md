# API Endpoints Contract

> This document defines the REST API contract. Frontend and Backend MUST follow these definitions.

## Base URL

- **Development**: `http://localhost:8080/api/v1`
- **Production**: `https://api.goplan.io/api/v1`

## Authentication

All endpoints except `/auth/*` require authentication via Bearer token.

```
Authorization: Bearer <access_token>
```

## Common Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes* | Bearer token (*except auth endpoints) |
| `Content-Type` | Yes | `application/json` |
| `X-Request-ID` | No | Client-generated UUID for tracing |
| `X-Idempotency-Key` | No | UUID for safe retries on POST/PUT/PATCH |

## Common Response Format

### Success Response
```json
{
  "data": { ... },
  "meta": {
    "requestId": "uuid"
  }
}
```

### List Response (Paginated)
```json
{
  "data": [ ... ],
  "meta": {
    "requestId": "uuid",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "totalPages": 5,
      "hasNext": true,
      "hasPrev": false
    }
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human readable message",
    "details": [
      { "field": "email", "message": "Invalid email format" }
    ]
  },
  "meta": {
    "requestId": "uuid"
  }
}
```

## Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `VALIDATION_ERROR` | Invalid request body |
| 400 | `BAD_REQUEST` | Malformed request |
| 401 | `UNAUTHORIZED` | Missing or invalid token |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Resource already exists |
| 422 | `UNPROCESSABLE_ENTITY` | Business logic error |
| 429 | `RATE_LIMITED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |

---

## Authentication Endpoints

### POST /auth/signup
Create a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "name": "John Doe"
}
```

**Response (201):**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "avatarUrl": null,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    },
    "accessToken": "jwt_token",
    "refreshToken": "refresh_token",
    "expiresAt": "2024-01-01T01:00:00Z"
  }
}
```

### POST /auth/login
Authenticate user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response (200):**
```json
{
  "data": {
    "user": { ... },
    "accessToken": "jwt_token",
    "refreshToken": "refresh_token",
    "expiresAt": "2024-01-01T01:00:00Z"
  }
}
```

### POST /auth/refresh
Refresh access token.

**Request:**
```json
{
  "refreshToken": "refresh_token"
}
```

**Response (200):**
```json
{
  "data": {
    "accessToken": "new_jwt_token",
    "refreshToken": "new_refresh_token",
    "expiresAt": "2024-01-01T01:00:00Z"
  }
}
```

### POST /auth/logout
Invalidate tokens.

**Request:**
```json
{
  "refreshToken": "refresh_token"
}
```

**Response (204):** No content

### POST /auth/forgot-password
Request password reset.

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Response (202):** Accepted (always, for security)

### POST /auth/reset-password
Reset password with token.

**Request:**
```json
{
  "token": "reset_token_from_email",
  "password": "newSecurePassword123"
}
```

**Response (200):** Success

---

## User Endpoints

### GET /users/me
Get current user profile.

**Response (200):**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "avatarUrl": "https://...",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

### PATCH /users/me
Update current user profile.

**Request:**
```json
{
  "name": "John Updated",
  "avatarUrl": "https://..."
}
```

### GET /users/me/workspaces
List workspaces user belongs to.

**Response (200):**
```json
{
  "data": [
    {
      "workspace": { ... },
      "role": "owner",
      "joinedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

## Workspace Endpoints

### POST /workspaces
Create a new workspace.

**Request:**
```json
{
  "name": "My Team",
  "slug": "my-team"
}
```

**Response (201):**
```json
{
  "data": {
    "id": "uuid",
    "name": "My Team",
    "slug": "my-team",
    "ownerId": "user_uuid",
    "settings": {
      "defaultView": "kanban",
      "aiEnabled": true
    },
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

### GET /workspaces/:workspaceId
Get workspace details.

### PATCH /workspaces/:workspaceId
Update workspace.

**Request:**
```json
{
  "name": "Updated Name",
  "settings": {
    "defaultView": "list",
    "aiEnabled": false
  }
}
```

### DELETE /workspaces/:workspaceId
Delete workspace (owner only).

### GET /workspaces/:workspaceId/members
List workspace members.

**Response (200):**
```json
{
  "data": [
    {
      "user": { ... },
      "role": "admin",
      "joinedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### POST /workspaces/:workspaceId/members
Invite member to workspace.

**Request:**
```json
{
  "email": "newmember@example.com",
  "role": "member"
}
```

### PATCH /workspaces/:workspaceId/members/:userId
Update member role.

**Request:**
```json
{
  "role": "admin"
}
```

### DELETE /workspaces/:workspaceId/members/:userId
Remove member from workspace.

---

## Plan Endpoints

### POST /workspaces/:workspaceId/plans
Create a new plan.

**Request:**
```json
{
  "name": "Q2 Product Launch",
  "description": "Launch our new product in Q2",
  "domain": "software",
  "startDate": "2024-04-01",
  "endDate": "2024-06-30",
  "customStatuses": [
    { "id": "backlog", "name": "Backlog", "color": "#6b7280", "order": 0, "isDefault": true, "isDone": false },
    { "id": "in_progress", "name": "In Progress", "color": "#3b82f6", "order": 1, "isDefault": false, "isDone": false },
    { "id": "review", "name": "Review", "color": "#f59e0b", "order": 2, "isDefault": false, "isDone": false },
    { "id": "done", "name": "Done", "color": "#10b981", "order": 3, "isDefault": false, "isDone": true }
  ],
  "tags": ["product", "launch"]
}
```

**Response (201):**
```json
{
  "data": {
    "id": "uuid",
    "workspaceId": "workspace_uuid",
    "name": "Q2 Product Launch",
    "description": "Launch our new product in Q2",
    "domain": "software",
    "status": "draft",
    "ownerId": "user_uuid",
    "startDate": "2024-04-01",
    "endDate": "2024-06-30",
    "customStatuses": [ ... ],
    "customFields": [],
    "tags": ["product", "launch"],
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

### GET /workspaces/:workspaceId/plans
List plans in workspace.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string | Filter by status |
| `domain` | string | Filter by domain |
| `ownerId` | string | Filter by owner |
| `search` | string | Search in name/description |
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 20, max: 100) |
| `sort` | string | Sort field (default: `-createdAt`) |

### GET /plans/:planId
Get plan details.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `include` | string | Comma-separated: `phases,tasks,milestones,stats` |

**Response with includes (200):**
```json
{
  "data": {
    "id": "uuid",
    "name": "Q2 Product Launch",
    ...
    "phases": [ ... ],
    "stats": {
      "totalTasks": 24,
      "completedTasks": 10,
      "overdueTasks": 2,
      "progress": 41.67
    }
  }
}
```

### PATCH /plans/:planId
Update plan.

### DELETE /plans/:planId
Delete plan.

### POST /plans/:planId/archive
Archive plan.

### POST /plans/:planId/activate
Activate plan (from draft).

---

## Phase Endpoints

### POST /plans/:planId/phases
Create a phase.

**Request:**
```json
{
  "name": "Research",
  "description": "Market research phase",
  "order": 0,
  "startDate": "2024-04-01",
  "endDate": "2024-04-15"
}
```

### GET /plans/:planId/phases
List phases in plan.

### PATCH /phases/:phaseId
Update phase.

### DELETE /phases/:phaseId
Delete phase (moves tasks to no phase).

### POST /phases/:phaseId/reorder
Reorder phases.

**Request:**
```json
{
  "order": 2
}
```

---

## Task Endpoints

### POST /plans/:planId/tasks
Create a task.

**Request:**
```json
{
  "title": "Design landing page",
  "description": "Create mockups for the new landing page",
  "phaseId": "phase_uuid",
  "status": "backlog",
  "priority": "high",
  "assigneeId": "user_uuid",
  "dueDate": "2024-04-10",
  "estimatedHours": 8,
  "tags": ["design", "frontend"],
  "customFieldValues": {
    "effort": "medium"
  }
}
```

### GET /plans/:planId/tasks
List tasks in plan.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string | Filter by status |
| `priority` | string | Filter by priority |
| `assigneeId` | string | Filter by assignee |
| `phaseId` | string | Filter by phase |
| `parentId` | string | Filter by parent (null for root tasks) |
| `dueBefore` | date | Due date before |
| `dueAfter` | date | Due date after |
| `search` | string | Search in title/description |
| `include` | string | Comma-separated: `subtasks,comments,assignee` |
| `page` | int | Page number |
| `limit` | int | Items per page |
| `sort` | string | Sort field |

### GET /tasks/:taskId
Get task details.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `include` | string | Comma-separated: `subtasks,comments,assignee,dependencies` |

### PATCH /tasks/:taskId
Update task.

**Request:**
```json
{
  "title": "Updated title",
  "status": "in_progress",
  "assigneeId": "new_user_uuid"
}
```

### DELETE /tasks/:taskId
Delete task (cascades to subtasks).

### POST /tasks/:taskId/subtasks
Create a subtask.

**Request:**
```json
{
  "title": "Subtask title",
  "priority": "medium"
}
```

### POST /tasks/:taskId/move
Move task to different status/phase.

**Request:**
```json
{
  "status": "done",
  "phaseId": "new_phase_uuid",
  "position": 0
}
```

### POST /tasks/:taskId/dependencies
Add task dependency.

**Request:**
```json
{
  "dependsOnId": "other_task_uuid"
}
```

### DELETE /tasks/:taskId/dependencies/:dependsOnId
Remove task dependency.

### POST /tasks/:taskId/assign
Assign task to user.

**Request:**
```json
{
  "assigneeId": "user_uuid"
}
```

### POST /tasks/bulk
Bulk create tasks.

**Request:**
```json
{
  "planId": "plan_uuid",
  "tasks": [
    { "title": "Task 1", "priority": "high" },
    { "title": "Task 2", "priority": "medium" }
  ]
}
```

### PATCH /tasks/bulk
Bulk update tasks.

**Request:**
```json
{
  "taskIds": ["task1_uuid", "task2_uuid"],
  "updates": {
    "status": "done"
  }
}
```

---

## Comment Endpoints

### POST /tasks/:taskId/comments
Add comment to task.

**Request:**
```json
{
  "content": "Great progress! @user_uuid can you review?",
  "mentions": ["user_uuid"]
}
```

### GET /tasks/:taskId/comments
List comments on task.

### PATCH /comments/:commentId
Update comment.

### DELETE /comments/:commentId
Delete comment.

---

## Milestone Endpoints

### POST /plans/:planId/milestones
Create milestone.

**Request:**
```json
{
  "name": "MVP Complete",
  "description": "Minimum viable product ready for testing",
  "dueDate": "2024-05-15",
  "linkedTaskIds": ["task1_uuid", "task2_uuid"]
}
```

### GET /plans/:planId/milestones
List milestones.

### PATCH /milestones/:milestoneId
Update milestone.

### DELETE /milestones/:milestoneId
Delete milestone.

---

## Activity & Search Endpoints

### GET /workspaces/:workspaceId/activity
Get workspace activity feed.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `planId` | string | Filter by plan |
| `userId` | string | Filter by user |
| `action` | string | Filter by action type |
| `since` | datetime | Activity since |
| `limit` | int | Items (default: 50) |

### GET /plans/:planId/activity
Get plan activity feed.

### GET /workspaces/:workspaceId/search
Search across workspace.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Search query |
| `type` | string | Filter by type: `plan`, `task`, `comment` |
| `limit` | int | Results limit |

---

## MCP / AI Endpoints

### POST /mcp/intent
Process natural language intent.

**Request:**
```json
{
  "input": "Create a plan for product launch in Q2",
  "context": {
    "workspaceId": "workspace_uuid",
    "planId": null
  }
}
```

**Response (200):**
```json
{
  "data": {
    "intentId": "uuid",
    "intentType": "CREATE_PLAN",
    "confidence": 0.94,
    "needsClarification": false,
    "entities": {
      "planName": "Product Launch",
      "planType": "software",
      "timeline": "Q2"
    },
    "draft": {
      "type": "plan",
      "data": {
        "name": "Product Launch",
        "domain": "software",
        ...
      }
    },
    "requiresApproval": true
  }
}
```

### POST /mcp/intent/:intentId/approve
Approve AI-generated action.

**Response (200):**
```json
{
  "data": {
    "result": {
      "type": "plan",
      "data": { ... }
    },
    "auditId": "uuid"
  }
}
```

### POST /mcp/intent/:intentId/reject
Reject AI-generated action.

**Request:**
```json
{
  "reason": "Not what I wanted"
}
```

### GET /mcp/tools
List available MCP tools.

**Response (200):**
```json
{
  "data": [
    {
      "name": "plan.create",
      "description": "Create a new plan",
      "parameters": { ... }
    },
    {
      "name": "task.create",
      "description": "Create a new task",
      "parameters": { ... }
    }
  ]
}
```

### POST /mcp/suggest
Get AI suggestions.

**Request:**
```json
{
  "type": "tasks",
  "context": {
    "planId": "plan_uuid"
  },
  "limit": 5
}
```

**Response (200):**
```json
{
  "data": {
    "suggestions": [
      {
        "title": "Create user documentation",
        "description": "Write end-user documentation for the new features",
        "priority": "medium",
        "confidence": 0.87
      }
    ]
  }
}
```

---

## WebSocket Events

### Connection
```javascript
const ws = new WebSocket('wss://api.goplan.io/ws?token=<access_token>');
```

### Subscribe to Plan
```json
{ "type": "subscribe", "channel": "plan:<planId>" }
```

### Events Received
```json
// Task updated
{
  "type": "task.updated",
  "planId": "uuid",
  "taskId": "uuid",
  "changes": { "status": "done" },
  "actor": { "id": "user_uuid", "name": "John" },
  "timestamp": "2024-01-01T00:00:00Z"
}

// User presence
{
  "type": "presence",
  "planId": "uuid",
  "users": [
    { "id": "uuid", "name": "John", "cursor": { "taskId": "uuid" } }
  ]
}

// New comment
{
  "type": "comment.added",
  "taskId": "uuid",
  "comment": { ... }
}
```

---

## Rate Limits

| Endpoint Category | Limit |
|-------------------|-------|
| Authentication | 10 req/min |
| Read operations | 100 req/min |
| Write operations | 30 req/min |
| AI/MCP operations | 20 req/min |
| Bulk operations | 5 req/min |

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704067200
```
