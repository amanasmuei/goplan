package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// contextKey is a type for context keys.
type contextKey string

const txKey contextKey = "pgx_tx"

// DBTX interface represents a database connection or transaction.
// Both pgxpool.Pool and pgx.Tx implement this interface.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TxManager provides transaction management.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager creates a new transaction manager.
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithTx executes fn within a transaction.
// If ctx already contains a transaction, it will be reused (nested call).
func (tm *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check for existing transaction
	if _, ok := ctx.Value(txKey).(pgx.Tx); ok {
		// Already in a transaction, just run the function
		return fn(ctx)
	}

	// Start new transaction
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Add transaction to context
	ctx = context.WithValue(ctx, txKey, tx)

	// Execute function
	if err := fn(ctx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetConn returns the transaction from context, or the pool if no transaction.
func GetConn(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return pool
}

// InTransaction returns true if the context contains a transaction.
func InTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(txKey).(pgx.Tx)
	return ok
}
