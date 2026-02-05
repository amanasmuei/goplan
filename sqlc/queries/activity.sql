-- name: CreateActivityLog :one
INSERT INTO activity_log (workspace_id, plan_id, task_id, user_id, action, details)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListActivityByWorkspace :many
SELECT al.*, u.name as user_name, u.email as user_email
FROM activity_log al
JOIN users u ON al.user_id = u.id
WHERE al.workspace_id = $1
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActivityByWorkspace :one
SELECT COUNT(*) FROM activity_log WHERE workspace_id = $1;

-- name: ListActivityByPlan :many
SELECT al.*, u.name as user_name, u.email as user_email
FROM activity_log al
JOIN users u ON al.user_id = u.id
WHERE al.plan_id = $1
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActivityByPlan :one
SELECT COUNT(*) FROM activity_log WHERE plan_id = $1;

-- name: ListActivityByTask :many
SELECT al.*, u.name as user_name, u.email as user_email
FROM activity_log al
JOIN users u ON al.user_id = u.id
WHERE al.task_id = $1
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActivityByTask :one
SELECT COUNT(*) FROM activity_log WHERE task_id = $1;

-- name: ListActivityFiltered :many
SELECT al.*, u.name as user_name, u.email as user_email
FROM activity_log al
JOIN users u ON al.user_id = u.id
WHERE (sqlc.narg('workspace_id')::uuid IS NULL OR al.workspace_id = sqlc.narg('workspace_id'))
  AND (sqlc.narg('plan_id')::uuid IS NULL OR al.plan_id = sqlc.narg('plan_id'))
  AND (sqlc.narg('task_id')::uuid IS NULL OR al.task_id = sqlc.narg('task_id'))
  AND (sqlc.narg('user_id')::uuid IS NULL OR al.user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('action')::varchar IS NULL OR al.action = sqlc.narg('action'))
  AND (sqlc.narg('since')::timestamptz IS NULL OR al.created_at >= sqlc.narg('since'))
  AND (sqlc.narg('until')::timestamptz IS NULL OR al.created_at <= sqlc.narg('until'))
ORDER BY al.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountActivityFiltered :one
SELECT COUNT(*) FROM activity_log al
WHERE (sqlc.narg('workspace_id')::uuid IS NULL OR al.workspace_id = sqlc.narg('workspace_id'))
  AND (sqlc.narg('plan_id')::uuid IS NULL OR al.plan_id = sqlc.narg('plan_id'))
  AND (sqlc.narg('task_id')::uuid IS NULL OR al.task_id = sqlc.narg('task_id'))
  AND (sqlc.narg('user_id')::uuid IS NULL OR al.user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('action')::varchar IS NULL OR al.action = sqlc.narg('action'))
  AND (sqlc.narg('since')::timestamptz IS NULL OR al.created_at >= sqlc.narg('since'))
  AND (sqlc.narg('until')::timestamptz IS NULL OR al.created_at <= sqlc.narg('until'));

-- name: GetRecentActivity :many
SELECT al.*, u.name as user_name, u.email as user_email
FROM activity_log al
JOIN users u ON al.user_id = u.id
WHERE al.user_id = $1
ORDER BY al.created_at DESC
LIMIT $2;
