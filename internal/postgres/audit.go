package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/mcp"
)

// AuditRepository implements repository.AuditRepository using PostgreSQL.
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// Create persists an audit record to the mcp_audit_log table.
func (r *AuditRepository) Create(ctx context.Context, record *mcp.AuditRecord) error {
	conn := GetConn(ctx, r.pool)

	var requestPayload, responsePayload []byte
	var err error

	// Build request payload from intent envelope and action args
	reqData := make(map[string]interface{})
	if record.IntentEnvelope != nil {
		reqData["intentEnvelope"] = record.IntentEnvelope
	}
	if record.ActionArgs != nil {
		reqData["actionArguments"] = record.ActionArgs
	}
	if len(reqData) > 0 {
		requestPayload, err = json.Marshal(reqData)
		if err != nil {
			return MapError(err, "mcp_audit_log")
		}
	}

	// Build response payload from result
	if record.Result != nil {
		responsePayload, err = json.Marshal(record.Result)
		if err != nil {
			return MapError(err, "mcp_audit_log")
		}
	}

	// Determine request_type from intent type or action tool
	requestType := "unknown"
	if record.IntentType != nil {
		requestType = *record.IntentType
	} else if record.ActionTool != nil {
		requestType = *record.ActionTool
	}

	_, err = conn.Exec(ctx,
		`INSERT INTO mcp_audit_log (workspace_id, user_id, request_type, request_payload, response_payload, status, error_message)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nullableUUID(record.WorkspaceID),
		uuidFromString(record.UserID),
		requestType,
		requestPayload,
		responsePayload,
		record.Status,
		record.ErrorMessage,
	)
	if err != nil {
		return MapError(err, "mcp_audit_log")
	}

	return nil
}

// nullableUUID returns a pgtype.UUID that is null if the string is empty.
func nullableUUID(s string) interface{} {
	if s == "" {
		return nil
	}
	return uuidFromString(s)
}
