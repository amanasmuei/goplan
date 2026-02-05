-- name: CreateComment :one
INSERT INTO comments (task_id, user_id, content, mentions)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCommentByID :one
SELECT c.*, u.name as user_name, u.email as user_email, u.avatar_url as user_avatar
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.id = $1;

-- name: UpdateComment :one
UPDATE comments
SET content = $2, mentions = $3
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;

-- name: ListCommentsByTask :many
SELECT c.*, u.name as user_name, u.email as user_email, u.avatar_url as user_avatar
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.task_id = $1
ORDER BY c.created_at ASC;

-- name: ListCommentsByTaskPaginated :many
SELECT c.*, u.name as user_name, u.email as user_email, u.avatar_url as user_avatar
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.task_id = $1
ORDER BY c.created_at ASC
LIMIT $2 OFFSET $3;

-- name: CountCommentsByTask :one
SELECT COUNT(*) FROM comments WHERE task_id = $1;
