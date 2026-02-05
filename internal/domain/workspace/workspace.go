// Package workspace provides the Workspace domain entity and related types.
package workspace

import (
	"regexp"
	"time"

	"github.com/goplan/goplan/internal/domain/shared"
)

// Workspace represents a workspace in the system.
type Workspace struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Slug      string            `json:"slug"`
	OwnerID   string            `json:"ownerId"`
	Settings  WorkspaceSettings `json:"settings"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// WorkspaceSettings represents the settings for a workspace.
type WorkspaceSettings struct {
	DefaultView string `json:"defaultView"`
	AIEnabled   bool   `json:"aiEnabled"`
}

// DefaultWorkspaceSettings returns the default workspace settings.
func DefaultWorkspaceSettings() WorkspaceSettings {
	return WorkspaceSettings{
		DefaultView: shared.ViewKanban,
		AIEnabled:   true,
	}
}

// WorkspaceMember represents a member of a workspace.
type WorkspaceMember struct {
	WorkspaceID string    `json:"workspaceId"`
	UserID      string    `json:"userId"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
}

// slugRegex is a regex for slug validation (lowercase alphanumeric + hyphens).
var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

// Validate validates the workspace fields.
func (w *Workspace) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	// Validate name
	if w.Name == "" {
		errs.Add("name", "name is required")
	} else if len(w.Name) > 255 {
		errs.Add("name", "name must be at most 255 characters")
	}

	// Validate slug
	if w.Slug == "" {
		errs.Add("slug", "slug is required")
	} else if len(w.Slug) < 3 || len(w.Slug) > 50 {
		errs.Add("slug", "slug must be between 3 and 50 characters")
	} else if !slugRegex.MatchString(w.Slug) {
		errs.Add("slug", "slug must be lowercase alphanumeric with hyphens only")
	}

	// Validate owner ID
	if w.OwnerID == "" {
		errs.Add("ownerId", "owner ID is required")
	}

	// Validate settings
	if !shared.IsValidDefaultView(w.Settings.DefaultView) {
		errs.Add("settings.defaultView", "invalid default view")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// NewWorkspace creates a new Workspace with the given parameters.
func NewWorkspace(id, name, slug, ownerID string, settings *WorkspaceSettings) *Workspace {
	now := time.Now().UTC()
	ws := &Workspace{
		ID:        id,
		Name:      name,
		Slug:      slug,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if settings != nil {
		ws.Settings = *settings
	} else {
		ws.Settings = DefaultWorkspaceSettings()
	}
	return ws
}

// NewWorkspaceMember creates a new WorkspaceMember.
func NewWorkspaceMember(workspaceID, userID, role string) *WorkspaceMember {
	return &WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
		JoinedAt:    time.Now().UTC(),
	}
}

// Validate validates the workspace member fields.
func (m *WorkspaceMember) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	if m.WorkspaceID == "" {
		errs.Add("workspaceId", "workspace ID is required")
	}

	if m.UserID == "" {
		errs.Add("userId", "user ID is required")
	}

	if !shared.IsValidMemberRole(m.Role) {
		errs.Add("role", "invalid role")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// CanManageMembers returns true if the member can manage other members.
func (m *WorkspaceMember) CanManageMembers() bool {
	return m.Role == shared.RoleOwner || m.Role == shared.RoleAdmin
}

// CanManagePlans returns true if the member can manage plans.
func (m *WorkspaceMember) CanManagePlans() bool {
	return m.Role == shared.RoleOwner || m.Role == shared.RoleAdmin || m.Role == shared.RoleMember
}

// CanViewOnly returns true if the member can only view.
func (m *WorkspaceMember) CanViewOnly() bool {
	return m.Role == shared.RoleViewer
}

// CreateWorkspaceInput represents the input for creating a new workspace.
type CreateWorkspaceInput struct {
	Name     string             `json:"name"`
	Slug     string             `json:"slug"`
	Settings *WorkspaceSettings `json:"settings,omitempty"`
}

// UpdateWorkspaceInput represents the input for updating a workspace.
type UpdateWorkspaceInput struct {
	Name     *string            `json:"name,omitempty"`
	Settings *WorkspaceSettings `json:"settings,omitempty"`
}

// AddMemberInput represents the input for adding a member to a workspace.
type AddMemberInput struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// WorkspaceResponse represents the workspace response for API.
type WorkspaceResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Slug      string            `json:"slug"`
	OwnerID   string            `json:"ownerId"`
	Settings  WorkspaceSettings `json:"settings"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
}

// ToResponse converts a Workspace to WorkspaceResponse with ISO 8601 timestamps.
func (w *Workspace) ToResponse() *WorkspaceResponse {
	return &WorkspaceResponse{
		ID:        w.ID,
		Name:      w.Name,
		Slug:      w.Slug,
		OwnerID:   w.OwnerID,
		Settings:  w.Settings,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
}

// WorkspaceMemberResponse represents the workspace member response for API.
type WorkspaceMemberResponse struct {
	WorkspaceID string `json:"workspaceId"`
	UserID      string `json:"userId"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joinedAt"`
}

// ToResponse converts a WorkspaceMember to WorkspaceMemberResponse.
func (m *WorkspaceMember) ToResponse() *WorkspaceMemberResponse {
	return &WorkspaceMemberResponse{
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Role:        m.Role,
		JoinedAt:    m.JoinedAt.Format(time.RFC3339),
	}
}
