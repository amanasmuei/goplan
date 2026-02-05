package models

import (
	"time"

	"github.com/google/uuid"
)

// TeamRole represents the role of a user within a team
type TeamRole string

const (
	TeamRoleOwner   TeamRole = "owner"
	TeamRoleManager TeamRole = "manager"
	TeamRoleMember  TeamRole = "member"
	TeamRoleViewer  TeamRole = "viewer"
)

// Team represents a team within an organization
type Team struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	CreatedBy      uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TeamID   uuid.UUID `json:"team_id" db:"team_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	Role     TeamRole  `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
	// Populated via JOIN
	User *User `json:"user,omitempty"`
}

// CreateTeamRequest is the request body for creating a new team
type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

// UpdateTeamRequest is the request body for updating a team
type UpdateTeamRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

// AddTeamMemberRequest is the request body for adding a member to a team
type AddTeamMemberRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Role   TeamRole  `json:"role" validate:"required,oneof=owner manager member viewer"`
}

// UpdateMemberRoleRequest is the request body for updating a member's role
type UpdateMemberRoleRequest struct {
	Role TeamRole `json:"role" validate:"required,oneof=owner manager member viewer"`
}

// TeamResponse is the response for a single team with additional data
type TeamResponse struct {
	Team        *Team          `json:"team"`
	MemberCount int            `json:"member_count"`
	Members     []TeamMember   `json:"members,omitempty"`
}

// TeamListResponse is the response for listing teams
type TeamListResponse struct {
	Teams []TeamResponse `json:"teams"`
	Total int64          `json:"total"`
}

// TeamMemberListResponse is the response for listing team members
type TeamMemberListResponse struct {
	Members []TeamMember `json:"members"`
	Total   int64        `json:"total"`
}

// CanManageTeam returns true if the role has permission to manage the team
func (r TeamRole) CanManageTeam() bool {
	return r == TeamRoleOwner || r == TeamRoleManager
}

// CanEditMembers returns true if the role has permission to add/remove members
func (r TeamRole) CanEditMembers() bool {
	return r == TeamRoleOwner || r == TeamRoleManager
}

// CanViewTeam returns true if the role has permission to view team details
func (r TeamRole) CanViewTeam() bool {
	return r == TeamRoleOwner || r == TeamRoleManager || r == TeamRoleMember || r == TeamRoleViewer
}

// IsValid returns true if the team role is a valid enum value
func (r TeamRole) IsValid() bool {
	switch r {
	case TeamRoleOwner, TeamRoleManager, TeamRoleMember, TeamRoleViewer:
		return true
	default:
		return false
	}
}
