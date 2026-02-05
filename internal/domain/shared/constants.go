// Package shared provides shared constants and types used across domain entities.
package shared

// Plan Domain Types
const (
	PlanDomainSoftware   = "software"
	PlanDomainEvent      = "event"
	PlanDomainOps        = "ops"
	PlanDomainCollection = "collection"
	PlanDomainGeneric    = "generic"
)

// PlanDomains is a list of all valid plan domains.
var PlanDomains = []string{
	PlanDomainSoftware,
	PlanDomainEvent,
	PlanDomainOps,
	PlanDomainCollection,
	PlanDomainGeneric,
}

// Plan Status Types
const (
	PlanStatusDraft     = "draft"
	PlanStatusActive    = "active"
	PlanStatusOnHold    = "on_hold"
	PlanStatusCompleted = "completed"
	PlanStatusArchived  = "archived"
)

// PlanStatuses is a list of all valid plan statuses.
var PlanStatuses = []string{
	PlanStatusDraft,
	PlanStatusActive,
	PlanStatusOnHold,
	PlanStatusCompleted,
	PlanStatusArchived,
}

// Task Priority Types
const (
	TaskPriorityLow      = "low"
	TaskPriorityMedium   = "medium"
	TaskPriorityHigh     = "high"
	TaskPriorityCritical = "critical"
)

// TaskPriorities is a list of all valid task priorities.
var TaskPriorities = []string{
	TaskPriorityLow,
	TaskPriorityMedium,
	TaskPriorityHigh,
	TaskPriorityCritical,
}

// Workspace Member Roles
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// MemberRoles is a list of all valid workspace member roles.
var MemberRoles = []string{
	RoleOwner,
	RoleAdmin,
	RoleMember,
	RoleViewer,
}

// Default View Types
const (
	ViewKanban   = "kanban"
	ViewList     = "list"
	ViewCalendar = "calendar"
	ViewGantt    = "gantt"
)

// DefaultViews is a list of all valid default view types.
var DefaultViews = []string{
	ViewKanban,
	ViewList,
	ViewCalendar,
	ViewGantt,
}

// Milestone Status Types
const (
	MilestoneStatusPending = "pending"
	MilestoneStatusReached = "reached"
	MilestoneStatusMissed  = "missed"
)

// MilestoneStatuses is a list of all valid milestone statuses.
var MilestoneStatuses = []string{
	MilestoneStatusPending,
	MilestoneStatusReached,
	MilestoneStatusMissed,
}

// Task Dependency Types
const (
	DependencyBlocks    = "blocks"
	DependencyBlockedBy = "blocked_by"
)

// DependencyTypes is a list of all valid dependency types.
var DependencyTypes = []string{
	DependencyBlocks,
	DependencyBlockedBy,
}

// Activity Action Types
const (
	ActivityPlanCreated    = "plan.created"
	ActivityPlanUpdated    = "plan.updated"
	ActivityPlanArchived   = "plan.archived"
	ActivityTaskCreated    = "task.created"
	ActivityTaskUpdated    = "task.updated"
	ActivityTaskDeleted    = "task.deleted"
	ActivityTaskMoved      = "task.moved"
	ActivityTaskAssigned   = "task.assigned"
	ActivityTaskCompleted  = "task.completed"
	ActivityCommentAdded   = "comment.added"
	ActivityMemberInvited  = "member.invited"
	ActivityMemberRemoved  = "member.removed"
)

// ActivityActions is a list of all valid activity actions.
var ActivityActions = []string{
	ActivityPlanCreated,
	ActivityPlanUpdated,
	ActivityPlanArchived,
	ActivityTaskCreated,
	ActivityTaskUpdated,
	ActivityTaskDeleted,
	ActivityTaskMoved,
	ActivityTaskAssigned,
	ActivityTaskCompleted,
	ActivityCommentAdded,
	ActivityMemberInvited,
	ActivityMemberRemoved,
}

// Field Types for Custom Fields
const (
	FieldTypeText        = "text"
	FieldTypeNumber      = "number"
	FieldTypeDate        = "date"
	FieldTypeSelect      = "select"
	FieldTypeMultiselect = "multiselect"
	FieldTypeCheckbox    = "checkbox"
)

// FieldTypes is a list of all valid field types.
var FieldTypes = []string{
	FieldTypeText,
	FieldTypeNumber,
	FieldTypeDate,
	FieldTypeSelect,
	FieldTypeMultiselect,
	FieldTypeCheckbox,
}

// IsValidPlanDomain checks if the given domain is valid.
func IsValidPlanDomain(domain string) bool {
	for _, d := range PlanDomains {
		if d == domain {
			return true
		}
	}
	return false
}

// IsValidPlanStatus checks if the given status is valid.
func IsValidPlanStatus(status string) bool {
	for _, s := range PlanStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidTaskPriority checks if the given priority is valid.
func IsValidTaskPriority(priority string) bool {
	for _, p := range TaskPriorities {
		if p == priority {
			return true
		}
	}
	return false
}

// IsValidMemberRole checks if the given role is valid.
func IsValidMemberRole(role string) bool {
	for _, r := range MemberRoles {
		if r == role {
			return true
		}
	}
	return false
}

// IsValidDefaultView checks if the given view is valid.
func IsValidDefaultView(view string) bool {
	for _, v := range DefaultViews {
		if v == view {
			return true
		}
	}
	return false
}

// IsValidFieldType checks if the given field type is valid.
func IsValidFieldType(fieldType string) bool {
	for _, f := range FieldTypes {
		if f == fieldType {
			return true
		}
	}
	return false
}
