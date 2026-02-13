# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoPlan is a universal, AI-powered project management platform with two Go backends and a React frontend:

1. **MCP Server** (root `cmd/server/`) — Go 1.24, `net/http`, `pgx/v5`, `sqlc`. Hosts the MCP intent/tool system, REST API v1, and Claude AI integration.
2. **Fiber API** (`backend/cmd/api/`) — Go + Fiber framework. REST API with Swagger docs, pgvector-powered semantic search, and an embedding service integration.
3. **Frontend** (`frontend/`) — React + TypeScript, Vite, TanStack Query, Zustand, Tailwind CSS.

Supporting services: PostgreSQL 15+ (with pgvector), Redis 7, Python FastAPI embedding service.

## Build Commands

### MCP Server (root)

```bash
go build ./...                             # build
go run ./cmd/server                        # run
go test ./...                              # test all
go test -v -race -cover ./...              # with race detection and coverage
go test -v ./internal/mcp/tools/...        # test specific package
go test -v -run TestTaskCreate ./...       # run specific test
golangci-lint run --timeout=5m ./...       # lint
golangci-lint run --fix --timeout=5m ./... # lint with auto-fix
cd sqlc && sqlc generate                   # generate type-safe SQL code
```

### Fiber API (backend/)

```bash
cd backend && go build -o bin/api ./cmd/api    # build
cd backend && go run ./cmd/api                 # run (requires DB)
cd backend && air                              # run with hot reload (.air.toml)
cd backend && go test -v -cover ./...          # test all
cd backend && go test -v ./internal/handlers/  # test one package
cd backend && go test -v -run TestFnName ./... # test one function
cd backend && golangci-lint run                # lint
cd backend && swag init -g cmd/api/main.go -o docs  # regenerate Swagger
```

### Frontend

```bash
cd frontend && yarn dev          # dev server (Vite, port 3000)
cd frontend && yarn build        # production build
cd frontend && yarn test         # unit tests (Vitest)
cd frontend && yarn test:e2e     # Playwright E2E (needs full stack)
cd frontend && yarn lint         # ESLint
cd frontend && yarn type-check   # TypeScript check
```

Package manager: **Yarn 4.12.0** (not npm) for the frontend.

### Database Migrations

```bash
# Root MCP server (golang-migrate)
migrate -path ./migrations -database "$DATABASE_URL" up
migrate -path ./migrations -database "$DATABASE_URL" down 1

# Fiber backend — auto-applied on postgres container start via init.sql
# Reset: docker-compose down -v && docker-compose up -d postgres
```

### Makefile (shortcuts)

```bash
make dev            # start postgres, redis, embedding-service via docker-compose
make api            # run Fiber backend locally
make frontend       # run frontend dev server
make build          # build backend + frontend
make test           # run all tests
make lint           # run linters
make docker-up      # start all Docker services
make docker-down    # stop all Docker services
make swagger        # regenerate Swagger docs
make install        # download Go + npm dependencies
```

## Architecture

### Two Backend Systems

The repo contains two separate Go modules with different approaches:

| | MCP Server (root) | Fiber API (backend/) |
|-|-|-|
| Entry point | `cmd/server/main.go` | `backend/cmd/api/main.go` |
| Framework | `net/http` | Go Fiber |
| DB access | `pgx/v5` + `sqlc` generated code | `pgx/v5` direct queries |
| Auth | JWT via `internal/auth/` | JWT via `internal/middleware/` |
| AI | Claude client + MCP intent routing | Embedding service for semantic search |
| go.mod | root `go.mod` (Go 1.24) | `backend/go.mod` |

### MCP Server Architecture (root)

All AI interactions flow through a standardized intent envelope system:

1. **Intent Envelope** (`MCPIntentEnvelope` in `internal/mcp/types.go`): Routes user intents with confidence scoring
2. **Agent Router** dispatches to specialized agents:
   - `PlannerAgent`: CREATE_PLAN, SUGGEST_TASKS
   - `ExecutorAgent`: ADD_TASK, UPDATE_TASK, MOVE_TASK, ASSIGN_TASK
   - `AnalystAgent`: ASK_SUMMARY (default fallback)
3. **Tool Registry** (`internal/mcp/registry.go`): Plugin system with 17+ registered MCP tools
4. **Audit System**: Full trail in `mcp_audit_log` table

**Confidence thresholds**:
- `>= 0.8` → proceed_with_confirmation
- `0.6–0.79` → proceed_with_uncertainty
- `< 0.6` → ask_clarification

**Key layers (root)**:
```
cmd/server/main.go           → Entry point, dependency injection, middleware chain
internal/api/handlers/       → REST API handlers
internal/api/middleware/      → Security headers
internal/auth/               → JWT auth + middleware
internal/mcp/server.go       → MCP HTTP endpoints
internal/mcp/tools/          → MCP tool implementations
internal/claude/             → Claude API client, service, safety checker
internal/domain/             → Domain entities (plan/, task/, user/, workspace/)
internal/postgres/           → Repository implementations (sqlc-generated)
internal/repository/         → Repository interfaces
internal/config/             → Config loading from env
```

**MCP Tool interface** (`internal/mcp/registry.go`):
```go
type Tool interface {
    Name() string        // e.g., "task.create"
    Description() string
    Execute(ctx context.Context, execCtx ExecutionContext, args map[string]interface{}) (interface{}, error)
}
```

**MCP endpoints**: `POST /mcp/intent`, `POST /mcp/tool/execute`, `POST /mcp/intent/approve`, `GET /mcp/tools`

Context passed via headers: `X-User-ID`, `X-Workspace-ID`, `X-User-Role`

### Fiber API Architecture (backend/)

```
backend/cmd/api/main.go          → Entry point, Fiber app, route registration
backend/internal/handlers/       → HTTP handlers (Swagger-annotated)
backend/internal/services/       → Business logic (TaskService with embedding client)
backend/internal/repository/     → PostgreSQL data access with pgvector
backend/internal/models/         → Domain structs, request/response types
backend/internal/middleware/     → JWT auth, request validation
backend/internal/workers/        → EmbeddingWorker (background, every 5 min)
backend/internal/config/         → Environment config
backend/internal/database/       → Connection pool (postgres.go)
```

Swagger docs available at `http://localhost:8080/swagger/index.html` when backend is running.

### Data Model

**MCP Server (root)** — Domain-Driven Design:
- **Plan** and **Task** are the only mandatory entities
- **Phases, milestones, dependencies, custom fields** are optional
- Status workflows customizable per plan via `custom_statuses` JSONB
- Plan domains: `software`, `event`, `ops`, `collection`, `generic`

**Fiber API (backend/)** — Task-centric with semantic features:
- Tasks with `description_embedding` (pgvector column) for similarity search
- Task links, justifications, blockers, reviews, acknowledgments
- Teams with hierarchical roles (owner, manager, member, viewer)
- Projects as logical groupings with team assignments

### Frontend Architecture

```
frontend/src/pages/         → Route components (TaskList, TaskDetail, Projects, Teams)
frontend/src/components/    → Reusable UI (Tailwind + lucide-react icons)
frontend/src/services/api.ts → Axios client with JWT interceptors
frontend/src/store/         → Zustand stores (authStore)
frontend/src/types/         → TypeScript types mirroring backend models
frontend/e2e/               → Playwright E2E tests
```

## Design Constraints

- **Human-in-the-loop**: All AI actions require user approval (draft-first pattern)
- **Full auditability**: Every action logged to `activity_log` and `mcp_audit_log`
- **Workspace isolation**: All data scoped to workspace with proper authorization
- **Agents communicate only via MCP envelopes**, no direct state mutation
- **Parameterized queries only** — use `$1`, `$2` etc. for all SQL

## Key Conventions

- Repository pattern: interfaces in `internal/repository/`, implementations in `internal/postgres/`
- Domain entities have `Validate()` methods returning `*ValidationErrors`
- MCP tools are namespaced: `workspace.*`, `plan.*`, `task.*`, `milestone.*`, `activity.*`
- Fiber handlers return `error` to Fiber; use `c.Status(code).JSON(ErrorResponse{})` for errors
- Auth context extracted via `getUserContext()` in Fiber handlers
- All sqlc queries live in `sqlc/queries/`; run `cd sqlc && sqlc generate` after editing
