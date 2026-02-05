// Package user provides the User domain entity and related types.
package user

import (
	"regexp"
	"time"

	"github.com/goplan/goplan/internal/domain/shared"
)

// User represents a user in the system.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UserWithPassword includes the password hash for authentication.
type UserWithPassword struct {
	User
	PasswordHash string `json:"-"`
}

// emailRegex is a simple regex for email validation.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validate validates the user fields.
func (u *User) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	// Validate email
	if u.Email == "" {
		errs.Add("email", "email is required")
	} else if len(u.Email) > 255 {
		errs.Add("email", "email must be at most 255 characters")
	} else if !emailRegex.MatchString(u.Email) {
		errs.Add("email", "email format is invalid")
	}

	// Validate name
	if u.Name == "" {
		errs.Add("name", "name is required")
	} else if len(u.Name) > 255 {
		errs.Add("name", "name must be at most 255 characters")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// NewUser creates a new User with the given parameters.
func NewUser(id, email, name string, avatarURL *string) *User {
	now := time.Now().UTC()
	return &User{
		ID:        id,
		Email:     email,
		Name:      name,
		AvatarURL: avatarURL,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the user's updatable fields.
func (u *User) Update(name string, avatarURL *string) {
	u.Name = name
	u.AvatarURL = avatarURL
	u.UpdatedAt = time.Now().UTC()
}

// CreateUserInput represents the input for creating a new user.
type CreateUserInput struct {
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Password  string  `json:"password"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}

// UpdateUserInput represents the input for updating a user.
type UpdateUserInput struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}

// UserResponse represents the user response for API.
type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// ToResponse converts a User to UserResponse with ISO 8601 timestamps.
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}
