# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoPlan is a universal project management and planning platform built as an MCP (Model Context Protocol) server in Go. It provides AI-assisted planning with multi-agent support, enabling natural language interaction for project and task management across different domains (software, events, operations, marketing, etc.).

## Build Commands

```bash
# Build
go build ./...

# Test
go test ./...
go test -v ./...                           # verbose
go test -v -race -cover ./...              # with race detection and coverage
go test -v ./internal/mcp/tools/...        # test specific package
go test -v -run TestTaskCreate ./...       # run specific test

# Lint
golangci-lint run --timeout=5m ./...       # full lint
golangci-lint run --fix --timeout=5m ./... # with auto-fix
go fmt ./...
go vet ./...

# Code generation
cd sqlc && sqlc generate                   # generate type-safe SQL code

# Database migrations
migrate -path ./migrations -database "$DATABASE_URL" up      # run migrations
migrate -path ./migrations -database "$DATABASE_URL" down 1  # rollback one
```

### Quick Start with Makefile

```bash
make setup          # full project setup (install deps, start services, migrate)
make dev            # start PostgreSQL, Redis, Mailhog
make test-backend   # run Go tests with coverage
make lint-backend   # run golangci-lint
make migrate-up     # run database migrations
make sqlc           # generate sqlc code
```

## Architecture

### MCP-First Design

All AI interactions flow through a standardized intent envelope system:

1. **Intent Envelope** (`MCPIntentEnvelope` in `internal/mcp/types.go`): Routes user intents with confidence scoring
2. **Agent Router**: Dispatches intents to specialized agents:
   - `PlannerAgent`: CREATE_PLAN, SUGGEST_TASKS
   - `ExecutorAgent`: ADD_TASK, UPDATE_TASK, MOVE_TASK, ASSIGN_TASK
   - `AnalystAgent`: ASK_SUMMARY (default fallback)
3. **Tool Registry** (`internal/mcp/registry.go`): Plugin system with 17+ registered MCP tools
4. **Audit System**: Full trail of all intents and actions in `mcp_audit_log` table

### Confidence-Based Behavior

```
>= 0.8  → proceed_with_confirmation
0.6-0.79 → proceed_with_uncertainty
< 0.6   → ask_clarification
```

### Key Layers

```
cmd/server/main.go           → Entry point, dependency injection
internal/api/handlers/       → REST API handlers
internal/api/middleware/     → Auth, rate limiting, security
internal/mcp/server.go       → MCP HTTP endpoints (/mcp/intent, /mcp/tool/execute)
internal/mcp/tools/          → Tool implementations (plan_tools.go, task_tools.go, etc.)
internal/claude/             → Claude API client and service
internal/domain/             → Domain entities (Plan, Task, Workspace, User)
internal/postgres/           → Repository implementations using sqlc
internal/repository/         → Repository interfaces
```

### MCP Tool Interface

Tools implement this interface in `internal/mcp/registry.go`:

```go
type Tool interface {
    Name() string        // e.g., "task.create"
    Description() string
    Execute(ctx context.Context, execCtx ExecutionContext, args map[string]interface{}) (interface{}, error)
}
```

### HTTP Gateway

- `POST /mcp/intent` - Parse and route user intent
- `POST /mcp/tool/execute` - Execute registered MCP tools
- `POST /mcp/intent/approve` - Approve draft action
- `GET /mcp/tools` - List available tools

Context passed via headers: `X-User-ID`, `X-Workspace-ID`

### Data Model

- **Plan and Task** are the only mandatory entities
- **Phases, milestones, dependencies, custom fields** are optional
- Status workflows are customizable per plan via `custom_statuses` JSONB field
- Plan domains: `software`, `event`, `ops`, `collection`, `generic`

## Design Constraints

- **Human-in-the-loop**: All AI actions require user approval (draft-first pattern)
- **Full auditability**: Every action logged to `activity_log` and `mcp_audit_log`
- **Workspace isolation**: All data scoped to workspace with proper authorization
- **Agents communicate only via MCP envelopes**, no direct state mutation
