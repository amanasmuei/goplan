# GoPlan API Reference

## Overview

The GoPlan API is a RESTful API that allows you to programmatically interact with GoPlan. All endpoints require authentication and return JSON responses.

**Base URL:** `https://goplan.yourcompany.com/api/v1`

## Authentication

### Bearer Token Authentication

Include your API token in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_API_TOKEN" \
  https://goplan.yourcompany.com/api/v1/tasks
```

### Obtaining a Token

1. Log in to GoPlan
2. Go to Settings > API Tokens
3. Create a new token with appropriate permissions
4. Store the token securely (it won't be shown again)

## Response Format

All responses follow this structure:

**Success:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Description of what went wrong",
    "details": { ... }
  }
}
```

## Rate Limiting

- **Default:** 1000 requests per hour
- Rate limit headers included in responses:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Requests remaining
  - `X-RateLimit-Reset`: Unix timestamp when limit resets

---

## Endpoints

### Tasks

#### List Tasks

```http
GET /tasks
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status (pending, acknowledged, in_progress, completed, cancelled) |
| `assignee_id` | uuid | Filter by assignee |
| `priority` | string | Filter by priority (low, medium, high, urgent) |
| `due_before` | datetime | Tasks due before this date |
| `due_after` | datetime | Tasks due after this date |
| `page` | integer | Page number (default: 1) |
| `limit` | integer | Items per page (default: 20, max: 100) |

**Response:**
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "Implement user authentication",
        "description": "Add JWT-based auth to the API",
        "status": "in_progress",
        "priority": "high",
        "assignee_id": "user-uuid",
        "assignee_name": "John Doe",
        "estimated_hours": 4.5,
        "predicted_hours": 5.2,
        "actual_hours": null,
        "due_date": "2026-01-25T17:00:00Z",
        "created_at": "2026-01-20T10:00:00Z",
        "updated_at": "2026-01-20T14:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "total_pages": 3
    }
  }
}
```

#### Get Task

```http
GET /tasks/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Implement user authentication",
    "description": "Add JWT-based auth to the API",
    "status": "in_progress",
    "priority": "high",
    "assignee_id": "user-uuid",
    "assignee_name": "John Doe",
    "creator_id": "creator-uuid",
    "creator_name": "Jane Smith",
    "team_id": "team-uuid",
    "team_name": "Backend Team",
    "estimated_hours": 4.5,
    "predicted_hours": 5.2,
    "actual_hours": null,
    "due_date": "2026-01-25T17:00:00Z",
    "acknowledged_at": "2026-01-20T11:00:00Z",
    "started_at": "2026-01-20T14:00:00Z",
    "completed_at": null,
    "tags": ["backend", "auth"],
    "parent_id": null,
    "created_at": "2026-01-20T10:00:00Z",
    "updated_at": "2026-01-20T14:30:00Z",
    "prediction": {
      "hours": 5.2,
      "confidence": 0.78,
      "factors": ["similar_tasks", "user_history"]
    }
  }
}
```

#### Create Task

```http
POST /tasks
```

**Request Body:**
```json
{
  "title": "New feature implementation",
  "description": "Detailed description of the task",
  "priority": "medium",
  "assignee_id": "user-uuid",
  "team_id": "team-uuid",
  "estimated_hours": 8,
  "due_date": "2026-01-30T17:00:00Z",
  "tags": ["feature", "frontend"],
  "parent_id": null
}
```

**Response:** Returns the created task object.

#### Update Task

```http
PATCH /tasks/{id}
```

**Request Body (all fields optional):**
```json
{
  "title": "Updated title",
  "description": "Updated description",
  "priority": "high",
  "assignee_id": "new-user-uuid",
  "estimated_hours": 10,
  "due_date": "2026-02-01T17:00:00Z",
  "tags": ["updated", "tags"]
}
```

#### Acknowledge Task

```http
POST /tasks/{id}/acknowledge
```

**Request Body:**
```json
{
  "accepted_prediction": true,
  "adjusted_hours": null,
  "notes": "Looks good, starting tomorrow"
}
```

Or with adjustment:
```json
{
  "accepted_prediction": false,
  "adjusted_hours": 6.5,
  "notes": "This will take longer due to dependencies"
}
```

#### Update Task Status

```http
POST /tasks/{id}/transition
```

**Request Body:**
```json
{
  "status": "in_progress",
  "notes": "Starting work on this now"
}
```

Valid transitions:
- `pending` → `acknowledged`
- `acknowledged` → `in_progress`
- `in_progress` → `blocked`, `review`, `completed`
- `blocked` → `in_progress`
- `review` → `completed`, `in_progress`

#### Complete Task

```http
POST /tasks/{id}/complete
```

**Request Body:**
```json
{
  "actual_hours": 5.5,
  "notes": "Completed successfully",
  "variance_explanation": "Took slightly longer due to code review feedback"
}
```

#### Delete Task

```http
DELETE /tasks/{id}
```

Note: Only pending tasks can be deleted. Use status transition to `cancelled` for active tasks.

---

### Users

#### List Users

```http
GET /users
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `team_id` | uuid | Filter by team |
| `role` | string | Filter by role (user, team_lead, admin) |
| `search` | string | Search by name or email |

#### Get Current User

```http
GET /users/me
```

#### Get User

```http
GET /users/{id}
```

---

### Teams

#### List Teams

```http
GET /teams
```

#### Get Team

```http
GET /teams/{id}
```

#### Get Team Members

```http
GET /teams/{id}/members
```

#### Get Team Tasks

```http
GET /teams/{id}/tasks
```

---

### Analytics

#### Get User Analytics

```http
GET /analytics/users/{id}
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_date` | date | Start of period (default: 30 days ago) |
| `end_date` | date | End of period (default: today) |

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "user-uuid",
    "period": {
      "start": "2025-12-21",
      "end": "2026-01-20"
    },
    "metrics": {
      "tasks_completed": 23,
      "tasks_on_time": 21,
      "completion_rate": 0.91,
      "average_variance": 0.12,
      "total_hours_logged": 156.5,
      "prediction_accuracy": 0.85
    },
    "trends": {
      "velocity": [
        { "week": "2025-W52", "completed": 5 },
        { "week": "2026-W01", "completed": 6 },
        { "week": "2026-W02", "completed": 7 },
        { "week": "2026-W03", "completed": 5 }
      ]
    }
  }
}
```

#### Get Team Analytics

```http
GET /analytics/teams/{id}
```

---

### Predictions

#### Get Prediction for New Task

```http
POST /predictions/estimate
```

**Request Body:**
```json
{
  "title": "Task title",
  "description": "Task description",
  "assignee_id": "user-uuid",
  "priority": "medium",
  "tags": ["backend"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "predicted_hours": 4.5,
    "confidence": 0.72,
    "range": {
      "min": 3.0,
      "max": 6.5
    },
    "factors": [
      {
        "name": "similar_tasks",
        "weight": 0.4,
        "description": "Based on 12 similar completed tasks"
      },
      {
        "name": "user_velocity",
        "weight": 0.35,
        "description": "User's historical completion rate"
      },
      {
        "name": "complexity",
        "weight": 0.25,
        "description": "Estimated from task description"
      }
    ]
  }
}
```

---

## Webhooks

GoPlan can send webhooks for various events.

### Event Types

| Event | Description |
|-------|-------------|
| `task.created` | New task created |
| `task.updated` | Task details changed |
| `task.acknowledged` | Task acknowledged by assignee |
| `task.started` | Task moved to in_progress |
| `task.completed` | Task completed |
| `task.cancelled` | Task cancelled |
| `task.blocked` | Task marked as blocked |

### Webhook Payload

```json
{
  "event": "task.completed",
  "timestamp": "2026-01-20T15:30:00Z",
  "data": {
    "task": {
      "id": "task-uuid",
      "title": "Task title",
      "status": "completed",
      "actual_hours": 5.5,
      "assignee_id": "user-uuid"
    },
    "previous_status": "review"
  }
}
```

### Webhook Verification

Verify webhooks using the signature header:

```
X-GoPlan-Signature: sha256=<signature>
```

Signature is HMAC-SHA256 of request body using your webhook secret.

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Invalid or missing auth token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `INVALID_TRANSITION` | 400 | Invalid status transition |
| `RATE_LIMITED` | 429 | Too many requests |
| `SERVER_ERROR` | 500 | Internal server error |

---

## SDK Examples

### JavaScript/TypeScript

```typescript
import { GoPlanClient } from '@goplan/sdk';

const client = new GoPlanClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://goplan.yourcompany.com/api/v1'
});

// List tasks
const tasks = await client.tasks.list({
  status: 'in_progress',
  assignee_id: 'user-uuid'
});

// Create a task
const task = await client.tasks.create({
  title: 'New feature',
  description: 'Description here',
  priority: 'high',
  assignee_id: 'user-uuid'
});

// Complete a task
await client.tasks.complete(task.id, {
  actual_hours: 4.5,
  notes: 'Done!'
});
```

### Python

```python
from goplan import GoPlanClient

client = GoPlanClient(
    api_key='your-api-key',
    base_url='https://goplan.yourcompany.com/api/v1'
)

# List tasks
tasks = client.tasks.list(
    status='in_progress',
    assignee_id='user-uuid'
)

# Create a task
task = client.tasks.create(
    title='New feature',
    description='Description here',
    priority='high',
    assignee_id='user-uuid'
)

# Complete a task
client.tasks.complete(
    task_id=task['id'],
    actual_hours=4.5,
    notes='Done!'
)
```

---

*API Version: 1.0*
