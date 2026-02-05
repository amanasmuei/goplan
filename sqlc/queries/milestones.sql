-- name: CreateMilestone :one
INSERT INTO milestones (plan_id, name, description, due_date, status, linked_task_ids)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMilestoneByID :one
SELECT * FROM milestones WHERE id = $1;

-- name: UpdateMilestone :one
UPDATE milestones
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    status = COALESCE(sqlc.narg('status'), status),
    linked_task_ids = COALESCE(sqlc.narg('linked_task_ids'), linked_task_ids)
WHERE id = $1
RETURNING *;

-- name: DeleteMilestone :exec
DELETE FROM milestones WHERE id = $1;

-- name: ListMilestonesByPlan :many
SELECT * FROM milestones WHERE plan_id = $1 ORDER BY due_date ASC;

-- name: CountMilestonesByPlan :one
SELECT COUNT(*) FROM milestones WHERE plan_id = $1;

-- name: LinkTasksToMilestone :one
UPDATE milestones
SET linked_task_ids = $2
WHERE id = $1
RETURNING *;

-- name: GetUpcomingMilestones :many
SELECT * FROM milestones
WHERE plan_id = $1
  AND due_date >= CURRENT_DATE
  AND status != 'completed'
ORDER BY due_date ASC
LIMIT $2;

-- name: GetOverdueMilestones :many
SELECT * FROM milestones
WHERE plan_id = $1
  AND due_date < CURRENT_DATE
  AND status != 'completed'
ORDER BY due_date ASC;

-- name: UpdateMilestoneStatus :one
UPDATE milestones SET status = $2 WHERE id = $1 RETURNING *;
