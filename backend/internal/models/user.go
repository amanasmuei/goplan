package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleTeamLead UserRole = "team_lead"
	UserRoleMember   UserRole = "member"
)

type User struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	Name           string    `json:"name" db:"name"`
	Role           UserRole  `json:"role" db:"role"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type Organization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Note: Project model is now in project.go with additional fields

type AuthClaims struct {
	UserID         uuid.UUID `json:"user_id"`
	Email          string    `json:"email"`
	Role           UserRole  `json:"role"`
	OrganizationID uuid.UUID `json:"organization_id"`
}
