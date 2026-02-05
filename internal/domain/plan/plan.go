// Package plan provides the Plan domain entity and related types.
package plan

import (
	"time"

	"github.com/goplan/goplan/internal/domain/shared"
)

// Plan represents a plan in the system.
type Plan struct {
	ID             string             `json:"id"`
	WorkspaceID    string             `json:"workspaceId"`
	Name           string             `json:"name"`
	Description    *string            `json:"description,omitempty"`
	Domain         string             `json:"domain"`
	Status         string             `json:"status"`
	OwnerID        string             `json:"ownerId"`
	StartDate      *string            `json:"startDate,omitempty"`
	EndDate        *string            `json:"endDate,omitempty"`
	CustomStatuses []StatusDefinition `json:"customStatuses"`
	CustomFields   []FieldDefinition  `json:"customFields"`
	Tags           []string           `json:"tags"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

// StatusDefinition represents a custom status for tasks in a plan.
type StatusDefinition struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Order     int    `json:"order"`
	IsDefault bool   `json:"isDefault"`
	IsDone    bool   `json:"isDone"`
}

// FieldDefinition represents a custom field for tasks in a plan.
type FieldDefinition struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
}

// DefaultStatuses returns the default status definitions for a new plan.
func DefaultStatuses() []StatusDefinition {
	return []StatusDefinition{
		{
			ID:        "todo",
			Name:      "To Do",
			Color:     "#6b7280",
			Order:     0,
			IsDefault: true,
			IsDone:    false,
		},
		{
			ID:        "in_progress",
			Name:      "In Progress",
			Color:     "#3b82f6",
			Order:     1,
			IsDefault: false,
			IsDone:    false,
		},
		{
			ID:        "done",
			Name:      "Done",
			Color:     "#10b981",
			Order:     2,
			IsDefault: false,
			IsDone:    true,
		},
	}
}

// Validate validates the plan fields.
func (p *Plan) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	// Validate name
	if p.Name == "" {
		errs.Add("name", "name is required")
	} else if len(p.Name) > 255 {
		errs.Add("name", "name must be at most 255 characters")
	}

	// Validate workspace ID
	if p.WorkspaceID == "" {
		errs.Add("workspaceId", "workspace ID is required")
	}

	// Validate owner ID
	if p.OwnerID == "" {
		errs.Add("ownerId", "owner ID is required")
	}

	// Validate domain
	if !shared.IsValidPlanDomain(p.Domain) {
		errs.Add("domain", "invalid plan domain")
	}

	// Validate status
	if !shared.IsValidPlanStatus(p.Status) {
		errs.Add("status", "invalid plan status")
	}

	// Validate custom statuses
	if len(p.CustomStatuses) == 0 {
		errs.Add("customStatuses", "at least one status is required")
	} else {
		hasDefault := false
		hasDone := false
		for _, s := range p.CustomStatuses {
			if s.IsDefault {
				hasDefault = true
			}
			if s.IsDone {
				hasDone = true
			}
		}
		if !hasDefault {
			errs.Add("customStatuses", "at least one status must be marked as default")
		}
		if !hasDone {
			errs.Add("customStatuses", "at least one status must be marked as done")
		}
	}

	// Validate custom fields
	for i, f := range p.CustomFields {
		if f.Name == "" {
			errs.Add("customFields", "field name is required")
		}
		if !shared.IsValidFieldType(f.Type) {
			errs.Add("customFields", "invalid field type at index "+string(rune(i)))
		}
		if (f.Type == shared.FieldTypeSelect || f.Type == shared.FieldTypeMultiselect) && len(f.Options) == 0 {
			errs.Add("customFields", "options are required for select/multiselect fields")
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// NewPlan creates a new Plan with the given parameters.
func NewPlan(id, workspaceID, name, ownerID string, domain string, description *string) *Plan {
	now := time.Now().UTC()
	return &Plan{
		ID:             id,
		WorkspaceID:    workspaceID,
		Name:           name,
		Description:    description,
		Domain:         domain,
		Status:         shared.PlanStatusDraft,
		OwnerID:        ownerID,
		CustomStatuses: DefaultStatuses(),
		CustomFields:   []FieldDefinition{},
		Tags:           []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// GetDefaultStatus returns the default status for new tasks.
func (p *Plan) GetDefaultStatus() string {
	for _, s := range p.CustomStatuses {
		if s.IsDefault {
			return s.ID
		}
	}
	if len(p.CustomStatuses) > 0 {
		return p.CustomStatuses[0].ID
	}
	return "todo"
}

// IsStatusDone returns true if the given status is marked as done.
func (p *Plan) IsStatusDone(statusID string) bool {
	for _, s := range p.CustomStatuses {
		if s.ID == statusID {
			return s.IsDone
		}
	}
	return false
}

// HasStatus returns true if the plan has the given status.
func (p *Plan) HasStatus(statusID string) bool {
	for _, s := range p.CustomStatuses {
		if s.ID == statusID {
			return true
		}
	}
	return false
}

// Phase represents a phase within a plan.
type Phase struct {
	ID          string    `json:"id"`
	PlanID      string    `json:"planId"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Order       int       `json:"order"`
	StartDate   *string   `json:"startDate,omitempty"`
	EndDate     *string   `json:"endDate,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Validate validates the phase fields.
func (ph *Phase) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	if ph.Name == "" {
		errs.Add("name", "name is required")
	} else if len(ph.Name) > 255 {
		errs.Add("name", "name must be at most 255 characters")
	}

	if ph.PlanID == "" {
		errs.Add("planId", "plan ID is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// NewPhase creates a new Phase with the given parameters.
func NewPhase(id, planID, name string, order int, description *string) *Phase {
	now := time.Now().UTC()
	return &Phase{
		ID:          id,
		PlanID:      planID,
		Name:        name,
		Description: description,
		Order:       order,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Milestone represents a milestone in a plan.
type Milestone struct {
	ID            string    `json:"id"`
	PlanID        string    `json:"planId"`
	Name          string    `json:"name"`
	Description   *string   `json:"description,omitempty"`
	DueDate       string    `json:"dueDate"`
	Status        string    `json:"status"`
	LinkedTaskIDs []string  `json:"linkedTaskIds"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Validate validates the milestone fields.
func (m *Milestone) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	if m.Name == "" {
		errs.Add("name", "name is required")
	} else if len(m.Name) > 255 {
		errs.Add("name", "name must be at most 255 characters")
	}

	if m.PlanID == "" {
		errs.Add("planId", "plan ID is required")
	}

	if m.DueDate == "" {
		errs.Add("dueDate", "due date is required")
	}

	validStatuses := map[string]bool{
		shared.MilestoneStatusPending: true,
		shared.MilestoneStatusReached: true,
		shared.MilestoneStatusMissed:  true,
	}
	if !validStatuses[m.Status] {
		errs.Add("status", "invalid milestone status")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// NewMilestone creates a new Milestone with the given parameters.
func NewMilestone(id, planID, name, dueDate string, description *string) *Milestone {
	now := time.Now().UTC()
	return &Milestone{
		ID:            id,
		PlanID:        planID,
		Name:          name,
		Description:   description,
		DueDate:       dueDate,
		Status:        shared.MilestoneStatusPending,
		LinkedTaskIDs: []string{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// CreatePlanInput represents the input for creating a new plan.
type CreatePlanInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Domain      string  `json:"domain"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
}

// UpdatePlanInput represents the input for updating a plan.
type UpdatePlanInput struct {
	Name           *string             `json:"name,omitempty"`
	Description    *string             `json:"description,omitempty"`
	Status         *string             `json:"status,omitempty"`
	StartDate      *string             `json:"startDate,omitempty"`
	EndDate        *string             `json:"endDate,omitempty"`
	CustomStatuses []StatusDefinition  `json:"customStatuses,omitempty"`
	CustomFields   []FieldDefinition   `json:"customFields,omitempty"`
	Tags           []string            `json:"tags,omitempty"`
}

// PlanResponse represents the plan response for API.
type PlanResponse struct {
	ID             string             `json:"id"`
	WorkspaceID    string             `json:"workspaceId"`
	Name           string             `json:"name"`
	Description    *string            `json:"description,omitempty"`
	Domain         string             `json:"domain"`
	Status         string             `json:"status"`
	OwnerID        string             `json:"ownerId"`
	StartDate      *string            `json:"startDate,omitempty"`
	EndDate        *string            `json:"endDate,omitempty"`
	CustomStatuses []StatusDefinition `json:"customStatuses"`
	CustomFields   []FieldDefinition  `json:"customFields"`
	Tags           []string           `json:"tags"`
	CreatedAt      string             `json:"createdAt"`
	UpdatedAt      string             `json:"updatedAt"`
}

// ToResponse converts a Plan to PlanResponse with ISO 8601 timestamps.
func (p *Plan) ToResponse() *PlanResponse {
	return &PlanResponse{
		ID:             p.ID,
		WorkspaceID:    p.WorkspaceID,
		Name:           p.Name,
		Description:    p.Description,
		Domain:         p.Domain,
		Status:         p.Status,
		OwnerID:        p.OwnerID,
		StartDate:      p.StartDate,
		EndDate:        p.EndDate,
		CustomStatuses: p.CustomStatuses,
		CustomFields:   p.CustomFields,
		Tags:           p.Tags,
		CreatedAt:      p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      p.UpdatedAt.Format(time.RFC3339),
	}
}

// PhaseResponse represents the phase response for API.
type PhaseResponse struct {
	ID          string  `json:"id"`
	PlanID      string  `json:"planId"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Order       int     `json:"order"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// ToResponse converts a Phase to PhaseResponse.
func (ph *Phase) ToResponse() *PhaseResponse {
	return &PhaseResponse{
		ID:          ph.ID,
		PlanID:      ph.PlanID,
		Name:        ph.Name,
		Description: ph.Description,
		Order:       ph.Order,
		StartDate:   ph.StartDate,
		EndDate:     ph.EndDate,
		CreatedAt:   ph.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   ph.UpdatedAt.Format(time.RFC3339),
	}
}

// MilestoneResponse represents the milestone response for API.
type MilestoneResponse struct {
	ID            string   `json:"id"`
	PlanID        string   `json:"planId"`
	Name          string   `json:"name"`
	Description   *string  `json:"description,omitempty"`
	DueDate       string   `json:"dueDate"`
	Status        string   `json:"status"`
	LinkedTaskIDs []string `json:"linkedTaskIds"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

// ToResponse converts a Milestone to MilestoneResponse.
func (m *Milestone) ToResponse() *MilestoneResponse {
	return &MilestoneResponse{
		ID:            m.ID,
		PlanID:        m.PlanID,
		Name:          m.Name,
		Description:   m.Description,
		DueDate:       m.DueDate,
		Status:        m.Status,
		LinkedTaskIDs: m.LinkedTaskIDs,
		CreatedAt:     m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     m.UpdatedAt.Format(time.RFC3339),
	}
}
