-- name: CreateTask :one
INSERT INTO tasks (
    plan_id, phase_id, parent_id, title, description, status,
    priority, assignee_id, due_date, estimated_hours,
    custom_field_values, tags, position
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetTaskByID :one
SELECT * FROM tasks WHERE id = $1;

-- name: UpdateTask :one
UPDATE tasks
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    priority = COALESCE(sqlc.narg('priority'), priority),
    phase_id = COALESCE(sqlc.narg('phase_id'), phase_id),
    assignee_id = COALESCE(sqlc.narg('assignee_id'), assignee_id),
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    estimated_hours = COALESCE(sqlc.narg('estimated_hours'), estimated_hours),
    custom_field_values = COALESCE(sqlc.narg('custom_field_values'), custom_field_values),
    tags = COALESCE(sqlc.narg('tags'), tags),
    position = COALESCE(sqlc.narg('position'), position)
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;

-- name: ListTasksByPlan :many
SELECT * FROM tasks
WHERE plan_id = $1
ORDER BY position ASC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTasksByPlanFiltered :many
SELECT * FROM tasks
WHERE plan_id = $1
  AND (sqlc.narg('phase_id')::uuid IS NULL OR phase_id = sqlc.narg('phase_id'))
  AND (sqlc.narg('parent_id')::uuid IS NULL OR parent_id = sqlc.narg('parent_id'))
  AND (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('priority')::varchar IS NULL OR priority = sqlc.narg('priority'))
  AND (sqlc.narg('assignee_id')::uuid IS NULL OR assignee_id = sqlc.narg('assignee_id'))
  AND (sqlc.narg('due_before')::date IS NULL OR due_date <= sqlc.narg('due_before'))
  AND (sqlc.narg('due_after')::date IS NULL OR due_date >= sqlc.narg('due_after'))
ORDER BY position ASC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTasksByPlan :one
SELECT COUNT(*) FROM tasks WHERE plan_id = $1;

-- name: CountTasksByPlanFiltered :one
SELECT COUNT(*) FROM tasks
WHERE plan_id = $1
  AND (sqlc.narg('phase_id')::uuid IS NULL OR phase_id = sqlc.narg('phase_id'))
  AND (sqlc.narg('parent_id')::uuid IS NULL OR parent_id = sqlc.narg('parent_id'))
  AND (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('priority')::varchar IS NULL OR priority = sqlc.narg('priority'))
  AND (sqlc.narg('assignee_id')::uuid IS NULL OR assignee_id = sqlc.narg('assignee_id'))
  AND (sqlc.narg('due_before')::date IS NULL OR due_date <= sqlc.narg('due_before'))
  AND (sqlc.narg('due_after')::date IS NULL OR due_date >= sqlc.narg('due_after'));

-- name: ListSubtasks :many
SELECT * FROM tasks WHERE parent_id = $1 ORDER BY position ASC;

-- name: SearchTasks :many
SELECT * FROM tasks
WHERE plan_id = $1
  AND to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, ''))
      @@ plainto_tsquery('english', $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountSearchTasks :one
SELECT COUNT(*) FROM tasks
WHERE plan_id = $1
  AND to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, ''))
      @@ plainto_tsquery('english', $2);

-- name: UpdateTaskStatus :one
UPDATE tasks SET status = $2 WHERE id = $1 RETURNING *;

-- name: UpdateTaskPosition :one
UPDATE tasks SET position = $2 WHERE id = $1 RETURNING *;

-- name: MoveTask :one
UPDATE tasks
SET phase_id = $2, position = $3
WHERE id = $1
RETURNING *;

-- name: GetMaxTaskPosition :one
SELECT COALESCE(MAX(position), -1)::integer FROM tasks WHERE plan_id = $1;

-- name: GetOverdueTasks :many
SELECT * FROM tasks
WHERE plan_id = $1
  AND due_date < CURRENT_DATE
  AND status NOT IN ('done', 'completed')
ORDER BY due_date ASC;

-- name: CountTasksByStatus :many
SELECT status, COUNT(*) as count
FROM tasks
WHERE plan_id = $1
GROUP BY status;

-- name: GetTasksForKanban :many
SELECT * FROM tasks
WHERE plan_id = $1 AND parent_id IS NULL
ORDER BY status, position ASC;

-- Add dependency
-- name: AddTaskDependency :exec
INSERT INTO task_dependencies (task_id, depends_on_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- Remove dependency
-- name: RemoveTaskDependency :exec
DELETE FROM task_dependencies
WHERE task_id = $1 AND depends_on_id = $2;

-- Get task dependencies (tasks this task depends on)
-- name: GetTaskDependencies :many
SELECT t.* FROM tasks t
JOIN task_dependencies td ON t.id = td.depends_on_id
WHERE td.task_id = $1;

-- Get tasks that depend on this task
-- name: GetTaskDependents :many
SELECT t.* FROM tasks t
JOIN task_dependencies td ON t.id = td.task_id
WHERE td.depends_on_id = $1;

-- Check if dependency would create a cycle
-- name: HasDependency :one
SELECT EXISTS(
    SELECT 1 FROM task_dependencies
    WHERE task_id = $1 AND depends_on_id = $2
);
