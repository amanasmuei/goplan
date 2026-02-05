package repository

import (
	"context"

	"github.com/goplan/goplan/internal/domain/workspace"
)

// WorkspaceFilterOptions defines workspace filtering options.
type WorkspaceFilterOptions struct {
	OwnerID  *string // Filter by owner
	MemberID *string // Filter by membership
}

// WorkspaceRepository defines workspace data access operations.
type WorkspaceRepository interface {
	// Create creates a new workspace and adds owner as member.
	Create(ctx context.Context, w *workspace.Workspace) error

	// GetByID retrieves a workspace by ID.
	GetByID(ctx context.Context, id string) (*workspace.Workspace, error)

	// GetBySlug retrieves a workspace by slug.
	GetBySlug(ctx context.Context, slug string) (*workspace.Workspace, error)

	// Update updates workspace fields.
	Update(ctx context.Context, id string, input *workspace.UpdateWorkspaceInput) (*workspace.Workspace, error)

	// Delete deletes a workspace by ID.
	Delete(ctx context.Context, id string) error

	// List retrieves workspaces with filtering and pagination.
	List(ctx context.Context, filter WorkspaceFilterOptions, pagination Pagination) (*PaginatedResult[workspace.Workspace], error)

	// ExistsBySlug checks if a workspace with the given slug exists.
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// AddMember adds a member to a workspace.
	AddMember(ctx context.Context, member *workspace.WorkspaceMember) error

	// UpdateMemberRole updates a member's role.
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error

	// RemoveMember removes a member from a workspace.
	RemoveMember(ctx context.Context, workspaceID, userID string) error

	// GetMember retrieves a specific member.
	GetMember(ctx context.Context, workspaceID, userID string) (*workspace.WorkspaceMember, error)

	// ListMembers retrieves all members of a workspace.
	ListMembers(ctx context.Context, workspaceID string) ([]*workspace.WorkspaceMember, error)

	// GetUserWorkspaces retrieves all workspaces a user is a member of.
	GetUserWorkspaces(ctx context.Context, userID string) ([]*workspace.Workspace, error)
}
