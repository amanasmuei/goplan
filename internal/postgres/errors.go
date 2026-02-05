package postgres

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/goplan/goplan/internal/domain/shared"
)

// PostgreSQL error codes.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgNotNullViolation    = "23502"
	pgCheckViolation      = "23514"
)

// MapError maps PostgreSQL errors to domain errors.
func MapError(err error, entityType string) error {
	if err == nil {
		return nil
	}

	// Check for no rows
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.NewNotFoundError(entityType, "")
	}

	// Check for PostgreSQL errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			field := extractFieldFromConstraint(pgErr.ConstraintName)
			return shared.NewAlreadyExistsError(entityType, field, "")
		case pgForeignKeyViolation:
			return shared.NewValidationError(
				extractFieldFromConstraint(pgErr.ConstraintName),
				"referenced entity does not exist",
			)
		case pgNotNullViolation:
			return shared.NewValidationError(pgErr.ColumnName, "cannot be null")
		case pgCheckViolation:
			return shared.NewValidationError(
				extractFieldFromConstraint(pgErr.ConstraintName),
				"constraint violation",
			)
		}
	}

	// Wrap unknown errors
	return shared.NewInternalError("database error", err)
}

// extractFieldFromConstraint extracts the field name from a constraint name.
// Common patterns: users_email_key, idx_users_email, fk_tasks_plan_id
func extractFieldFromConstraint(constraint string) string {
	if constraint == "" {
		return "unknown"
	}

	// Split by underscore
	parts := strings.Split(constraint, "_")
	if len(parts) < 2 {
		return constraint
	}

	// Try to find meaningful field name
	// Skip common prefixes: idx_, fk_, pk_, uq_
	for i, part := range parts {
		if part == "idx" || part == "fk" || part == "pk" || part == "uq" {
			continue
		}
		// Skip table name (usually first or second part)
		if i < 2 && isTableName(part) {
			continue
		}
		// Skip "key" suffix
		if part == "key" {
			continue
		}
		return part
	}

	// Fallback: return second-to-last part (usually field name)
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return constraint
}

// isTableName checks if a string looks like a table name.
func isTableName(s string) bool {
	tables := map[string]bool{
		"users":              true,
		"workspaces":         true,
		"workspace":          true,
		"plans":              true,
		"phases":             true,
		"tasks":              true,
		"milestones":         true,
		"comments":           true,
		"activity":           true,
		"mcp":                true,
		"ai":                 true,
		"members":            true,
		"dependencies":       true,
		"task_dependencies":  true,
		"workspace_members":  true,
	}
	return tables[s]
}

// Error codes used for comparison.
const (
	ErrCodeNotFound      = "ENTITY_NOT_FOUND"
	ErrCodeAlreadyExists = "ENTITY_ALREADY_EXISTS"
)

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *shared.DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code == ErrCodeNotFound
	}
	return errors.Is(err, pgx.ErrNoRows)
}

// IsAlreadyExists returns true if the error is a unique constraint violation.
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *shared.DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code == ErrCodeAlreadyExists
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgUniqueViolation
	}
	return false
}
