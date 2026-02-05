-- name: CreateWorkspace :one
INSERT INTO workspaces (name, slug, owner_id, settings)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWorkspaceByID :one
SELECT * FROM workspaces WHERE id = $1;

-- name: GetWorkspaceBySlug :one
SELECT * FROM workspaces WHERE slug = $1;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    settings = COALESCE(sqlc.narg('settings'), settings)
WHERE id = $1
RETURNING *;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces WHERE id = $1;

-- name: ListWorkspaces :many
SELECT * FROM workspaces
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountWorkspaces :one
SELECT COUNT(*) FROM workspaces;

-- name: AddWorkspaceMember :exec
INSERT INTO workspace_members (workspace_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = $3;

-- name: RemoveWorkspaceMember :exec
DELETE FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: UpdateMemberRole :exec
UPDATE workspace_members
SET role = $3
WHERE workspace_id = $1 AND user_id = $2;

-- name: GetWorkspaceMember :one
SELECT wm.workspace_id, wm.user_id, wm.role, wm.joined_at,
       u.email, u.name, u.avatar_url
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1 AND wm.user_id = $2;

-- name: ListWorkspaceMembers :many
SELECT wm.workspace_id, wm.user_id, wm.role, wm.joined_at,
       u.email, u.name, u.avatar_url
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.joined_at;

-- name: CountWorkspaceMembers :one
SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1;

-- name: GetUserWorkspaces :many
SELECT w.*, wm.role, wm.joined_at
FROM workspaces w
JOIN workspace_members wm ON w.id = wm.workspace_id
WHERE wm.user_id = $1
ORDER BY wm.joined_at DESC;

-- name: IsMember :one
SELECT EXISTS(
    SELECT 1 FROM workspace_members
    WHERE workspace_id = $1 AND user_id = $2
);

-- name: GetMemberRole :one
SELECT role FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: ExistsBySlug :one
SELECT EXISTS(SELECT 1 FROM workspaces WHERE slug = $1);
