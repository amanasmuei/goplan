-- name: CreatePhase :one
INSERT INTO phases (plan_id, name, description, "order", start_date, end_date)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPhaseByID :one
SELECT * FROM phases WHERE id = $1;

-- name: UpdatePhase :one
UPDATE phases
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    "order" = COALESCE(sqlc.narg('order'), "order"),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = COALESCE(sqlc.narg('end_date'), end_date)
WHERE id = $1
RETURNING *;

-- name: DeletePhase :exec
DELETE FROM phases WHERE id = $1;

-- name: ListPhasesByPlan :many
SELECT * FROM phases WHERE plan_id = $1 ORDER BY "order" ASC;

-- name: CountPhasesByPlan :one
SELECT COUNT(*) FROM phases WHERE plan_id = $1;

-- name: ReorderPhases :exec
UPDATE phases
SET "order" = data.new_order
FROM (
    SELECT unnest($1::uuid[]) AS phase_id,
           generate_series(0, array_length($1::uuid[], 1) - 1) AS new_order
) AS data
WHERE phases.id = data.phase_id;

-- name: GetMaxPhaseOrder :one
SELECT COALESCE(MAX("order"), -1)::integer FROM phases WHERE plan_id = $1;
