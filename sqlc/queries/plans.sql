-- name: CreatePlan :one
INSERT INTO plans (
    workspace_id, name, description, domain, status,
    owner_id, start_date, end_date, custom_statuses, custom_fields, tags
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetPlanByID :one
SELECT * FROM plans WHERE id = $1;

-- name: UpdatePlan :one
UPDATE plans
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    domain = COALESCE(sqlc.narg('domain'), domain),
    status = COALESCE(sqlc.narg('status'), status),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = COALESCE(sqlc.narg('end_date'), end_date),
    custom_statuses = COALESCE(sqlc.narg('custom_statuses'), custom_statuses),
    custom_fields = COALESCE(sqlc.narg('custom_fields'), custom_fields),
    tags = COALESCE(sqlc.narg('tags'), tags)
WHERE id = $1
RETURNING *;

-- name: UpdatePlanStatus :one
UPDATE plans SET status = $2 WHERE id = $1 RETURNING *;

-- name: DeletePlan :exec
DELETE FROM plans WHERE id = $1;

-- name: ListPlansByWorkspace :many
SELECT * FROM plans
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPlansByWorkspaceFiltered :many
SELECT * FROM plans
WHERE workspace_id = $1
  AND (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('owner_id')::uuid IS NULL OR owner_id = sqlc.narg('owner_id'))
  AND (sqlc.narg('domain')::varchar IS NULL OR domain = sqlc.narg('domain'))
ORDER BY
  CASE WHEN sqlc.narg('sort_field')::varchar = 'name' AND sqlc.narg('sort_order')::varchar = 'ASC' THEN name END ASC,
  CASE WHEN sqlc.narg('sort_field')::varchar = 'name' AND sqlc.narg('sort_order')::varchar = 'DESC' THEN name END DESC,
  CASE WHEN sqlc.narg('sort_field')::varchar = 'created_at' AND sqlc.narg('sort_order')::varchar = 'ASC' THEN created_at END ASC,
  CASE WHEN sqlc.narg('sort_field')::varchar = 'updated_at' AND sqlc.narg('sort_order')::varchar = 'ASC' THEN updated_at END ASC,
  CASE WHEN sqlc.narg('sort_field')::varchar = 'updated_at' AND sqlc.narg('sort_order')::varchar = 'DESC' THEN updated_at END DESC,
  created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPlansByWorkspace :one
SELECT COUNT(*) FROM plans WHERE workspace_id = $1;

-- name: CountPlansByWorkspaceFiltered :one
SELECT COUNT(*) FROM plans
WHERE workspace_id = $1
  AND (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('owner_id')::uuid IS NULL OR owner_id = sqlc.narg('owner_id'))
  AND (sqlc.narg('domain')::varchar IS NULL OR domain = sqlc.narg('domain'));

-- name: GetPlansByOwner :many
SELECT * FROM plans WHERE owner_id = $1 ORDER BY created_at DESC;
