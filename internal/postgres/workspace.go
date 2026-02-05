package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/domain/workspace"
	"github.com/goplan/goplan/internal/postgres/sqlc"
	"github.com/goplan/goplan/internal/repository"
)

// WorkspaceRepository implements repository.WorkspaceRepository using PostgreSQL.
type WorkspaceRepository struct {
	pool *pgxpool.Pool
}

// NewWorkspaceRepository creates a new WorkspaceRepository.
func NewWorkspaceRepository(pool *pgxpool.Pool) *WorkspaceRepository {
	return &WorkspaceRepository{pool: pool}
}

// Create creates a new workspace and adds owner as member.
func (r *WorkspaceRepository) Create(ctx context.Context, w *workspace.Workspace) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	settingsJSON, err := json.Marshal(w.Settings)
	if err != nil {
		return err
	}

	result, err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		Name:     w.Name,
		Slug:     w.Slug,
		OwnerID:  w.OwnerID,
		Settings: settingsJSON,
	})
	if err != nil {
		return MapError(err, "workspace")
	}

	w.ID = result.ID
	w.CreatedAt = result.CreatedAt.Time
	w.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// GetByID retrieves a workspace by ID.
func (r *WorkspaceRepository) GetByID(ctx context.Context, id string) (*workspace.Workspace, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetWorkspaceByID(ctx, id)
	if err != nil {
		return nil, MapError(err, "workspace")
	}

	return sqlcWorkspaceToDomain(result)
}

// GetBySlug retrieves a workspace by slug.
func (r *WorkspaceRepository) GetBySlug(ctx context.Context, slug string) (*workspace.Workspace, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		return nil, MapError(err, "workspace")
	}

	return sqlcWorkspaceToDomain(result)
}

// Update updates workspace fields.
func (r *WorkspaceRepository) Update(ctx context.Context, id string, input *workspace.UpdateWorkspaceInput) (*workspace.Workspace, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	params := sqlc.UpdateWorkspaceParams{
		ID: id,
	}
	if input.Name != nil {
		params.Name = textPtr(*input.Name)
	}
	if input.Settings != nil {
		settingsJSON, err := json.Marshal(input.Settings)
		if err != nil {
			return nil, err
		}
		params.Settings = settingsJSON
	}

	result, err := q.UpdateWorkspace(ctx, params)
	if err != nil {
		return nil, MapError(err, "workspace")
	}

	return sqlcWorkspaceToDomain(result)
}

// Delete deletes a workspace by ID.
func (r *WorkspaceRepository) Delete(ctx context.Context, id string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.DeleteWorkspace(ctx, id)
	if err != nil {
		return MapError(err, "workspace")
	}

	return nil
}

// List retrieves workspaces with filtering and pagination.
func (r *WorkspaceRepository) List(ctx context.Context, filter repository.WorkspaceFilterOptions, pagination repository.Pagination) (*repository.PaginatedResult[workspace.Workspace], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	workspaces, err := q.ListWorkspaces(ctx, sqlc.ListWorkspacesParams{
		Limit:  int32(pagination.PageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "workspace")
	}

	count, err := q.CountWorkspaces(ctx)
	if err != nil {
		return nil, MapError(err, "workspace")
	}

	items := make([]workspace.Workspace, len(workspaces))
	for i, ws := range workspaces {
		w, err := sqlcWorkspaceToDomain(ws)
		if err != nil {
			return nil, err
		}
		items[i] = *w
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// ExistsBySlug checks if a workspace with the given slug exists.
func (r *WorkspaceRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	exists, err := q.ExistsBySlug(ctx, slug)
	if err != nil {
		return false, MapError(err, "workspace")
	}

	return exists, nil
}

// AddMember adds a member to a workspace.
func (r *WorkspaceRepository) AddMember(ctx context.Context, member *workspace.WorkspaceMember) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.AddWorkspaceMember(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Role:        member.Role,
	})
	if err != nil {
		return MapError(err, "workspace_member")
	}

	return nil
}

// UpdateMemberRole updates a member's role.
func (r *WorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.UpdateMemberRole(ctx, sqlc.UpdateMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
	})
	if err != nil {
		return MapError(err, "workspace_member")
	}

	return nil
}

// RemoveMember removes a member from a workspace.
func (r *WorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.RemoveWorkspaceMember(ctx, sqlc.RemoveWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return MapError(err, "workspace_member")
	}

	return nil
}

// GetMember retrieves a specific member.
func (r *WorkspaceRepository) GetMember(ctx context.Context, workspaceID, userID string) (*workspace.WorkspaceMember, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return nil, MapError(err, "workspace_member")
	}

	return &workspace.WorkspaceMember{
		WorkspaceID: result.WorkspaceID,
		UserID:      result.UserID,
		Role:        result.Role,
		JoinedAt:    result.JoinedAt.Time,
	}, nil
}

// ListMembers retrieves all members of a workspace.
func (r *WorkspaceRepository) ListMembers(ctx context.Context, workspaceID string) ([]*workspace.WorkspaceMember, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	members, err := q.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, MapError(err, "workspace_member")
	}

	result := make([]*workspace.WorkspaceMember, len(members))
	for i, m := range members {
		result[i] = &workspace.WorkspaceMember{
			WorkspaceID: m.WorkspaceID,
			UserID:      m.UserID,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt.Time,
		}
	}

	return result, nil
}

// GetUserWorkspaces retrieves all workspaces a user is a member of.
func (r *WorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*workspace.Workspace, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	workspaces, err := q.GetUserWorkspaces(ctx, userID)
	if err != nil {
		return nil, MapError(err, "workspace")
	}

	result := make([]*workspace.Workspace, len(workspaces))
	for i, ws := range workspaces {
		var settings workspace.WorkspaceSettings
		if len(ws.Settings) > 0 {
			if err := json.Unmarshal(ws.Settings, &settings); err != nil {
				settings = workspace.DefaultWorkspaceSettings()
			}
		} else {
			settings = workspace.DefaultWorkspaceSettings()
		}

		result[i] = &workspace.Workspace{
			ID:        ws.ID,
			Name:      ws.Name,
			Slug:      ws.Slug,
			OwnerID:   ws.OwnerID,
			Settings:  settings,
			CreatedAt: ws.CreatedAt.Time,
			UpdatedAt: ws.UpdatedAt.Time,
		}
	}

	return result, nil
}

// Helper functions

func sqlcWorkspaceToDomain(ws sqlc.Workspace) (*workspace.Workspace, error) {
	var settings workspace.WorkspaceSettings
	if len(ws.Settings) > 0 {
		if err := json.Unmarshal(ws.Settings, &settings); err != nil {
			settings = workspace.DefaultWorkspaceSettings()
		}
	} else {
		settings = workspace.DefaultWorkspaceSettings()
	}

	return &workspace.Workspace{
		ID:        ws.ID,
		Name:      ws.Name,
		Slug:      ws.Slug,
		OwnerID:   ws.OwnerID,
		Settings:  settings,
		CreatedAt: ws.CreatedAt.Time,
		UpdatedAt: ws.UpdatedAt.Time,
	}, nil
}
