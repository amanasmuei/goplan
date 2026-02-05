// Package repository defines interfaces for data access operations.
package repository

import (
	"context"
)

// Pagination represents pagination parameters.
type Pagination struct {
	Page     int // 1-indexed page number
	PageSize int // Number of items per page
}

// DefaultPagination returns default pagination settings.
func DefaultPagination() Pagination {
	return Pagination{
		Page:     1,
		PageSize: 20,
	}
}

// Normalize ensures pagination values are within valid ranges.
func (p Pagination) Normalize() Pagination {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p
}

// Offset calculates the offset for SQL queries.
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PaginatedResult represents a paginated query result.
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"totalCount"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
}

// NewPaginatedResult creates a new paginated result.
func NewPaginatedResult[T any](items []T, totalCount int64, pagination Pagination) *PaginatedResult[T] {
	totalPages := int((totalCount + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}
	return &PaginatedResult[T]{
		Items:      items,
		TotalCount: totalCount,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}
}

// SortOrder represents sort direction.
type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

// TxManager provides transaction management.
type TxManager interface {
	// WithTx executes fn within a transaction.
	// If the context already contains a transaction, it will be reused.
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
