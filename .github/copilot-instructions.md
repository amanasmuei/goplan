# GoPlan - Copilot Instructions

GoPlan is a **planning-first task management system** with semantic intelligence using embeddings. It consists of three services: a Go/Fiber backend API, a React/TypeScript frontend, and a Python FastAPI embedding service.

## Build, Test, and Lint Commands

### Backend (Go)
- **Build**: `cd backend && go build -o bin/api ./cmd/api` or `make build`
- **Run locally**: `make api` (requires DB running)
- **Run with hot reload**: `cd backend && air` (uses `.air.toml` config)
- **Test all**: `cd backend && go test -v -cover ./...`
- **Test one package**: `cd backend && go test -v ./internal/handlers/`
- **Test one function**: `cd backend && go test -v -run TestFunctionName ./...`
- **Lint**: `cd backend && golangci-lint run`
- **Format**: `cd backend && go fmt ./...`
- **Generate Swagger docs**: `make swagger` or `cd backend && swag init -g cmd/api/main.go -o docs`

### Frontend (React + TypeScript)
- **Dev server**: `cd frontend && npm run dev` or `make frontend`
- **Build**: `cd frontend && npm run build` or `make build`
- **Type check**: `cd frontend && npm run type-check`
- **Test all**: `cd frontend && npm test` or `npm run test:watch` for watch mode
- **Test one file**: `cd frontend && npm test -- TaskList.test.tsx`
- **Lint**: `cd frontend && npm run lint`
- **Format**: `cd frontend && npm run format` or `prettier --write "src/**/*.{ts,tsx,css}"`
- **E2E tests**: `cd frontend && npm run test:e2e` (requires frontend & backend running)
- **E2E UI mode**: `cd frontend && npm run test:e2e:ui`
- **Coverage**: `cd frontend && npm run test:coverage`

### Embedding Service (Python)
- **Run**: Docker only (included in docker-compose)
- **Manual**: `python main.py` from `embedding-service/` (requires dependencies)

### Docker & Infrastructure
- **Start all services**: `make docker-up` or `docker-compose up -d`
- **Stop all**: `make docker-down` or `docker-compose down`
- **Rebuild and start**: `make docker-rebuild`
- **View logs**: `make docker-logs`
- **Database migrations**: Auto-run on postgres startup via `init.sql`

## High-Level Architecture

### System Overview
```
┌─────────────────────────────────────────────────────────────────┐
│  Frontend (React + TS)                                          │
│  - React Router for SPA navigation                              │
│  - TanStack React Query for data fetching & caching             │
│  - Zustand for auth/user state management                       │
│  - Tailwind CSS + custom components                             │
│  - Axios with JWT auth via interceptors                         │
└────────────┬────────────────────────────────────────────────────┘
             │ HTTP/JSON
             ▼
┌─────────────────────────────────────────────────────────────────┐
│  Backend API (Go + Fiber)                                       │
│  - REST API (v1) with Swagger docs at /swagger/index.html       │
│  - JWT Bearer token authentication                              │
│  - PostgreSQL with pgvector for semantic search                 │
│  - Architecture: Handlers → Services → Repositories             │
│  - CORS & JSON middleware                                       │
└────────────┬────────────┬────────────────────────────────────────┘
             │            │
             ▼            ▼
    ┌─────────────────┐  ┌──────────────────────────┐
    │  PostgreSQL     │  │  Embedding Service       │
    │  (pgvector)     │  │  (FastAPI + sentence-    │
    │  - Data storage │  │   transformers)          │
    │  - Vector       │  │  - Text embedding gen    │
    │    embeddings   │  │  - Model: all-MiniLM    │
    │  - Task links   │  │  - REST API on :8000    │
    │  - Reviews      │  │                          │
    │  - Blockers     │  │                          │
    └─────────────────┘  └──────────────────────────┘
```

### Backend Layering

**Handlers** (`internal/handlers/`)
- Receive HTTP requests, parse bodies, extract auth context
- Call services, return JSON responses
- Swagger-annotated for API documentation
- Each domain has its own handler: `auth_handler.go`, `task_handler.go`, `project_handler.go`, etc.

**Services** (`internal/services/`)
- Business logic and orchestration
- Call multiple repositories
- Manage cross-cutting concerns (validation, state transitions)
- `TaskService` integrates with embedding client for semantic features

**Repositories** (`internal/repository/`)
- Data access layer with PostgreSQL
- Use pgx connection pool
- Named by domain: `task_repository.go`, `project_repository.go`, etc.
- Vector operations via pgvector-go for semantic search

**Models** (`internal/models/`)
- Go structs for domain entities and API contracts
- Request/Response types with JSON tags
- Task statuses: Draft → Pending Acknowledgment → Acknowledged → Active → Blocked/Pending Review → Completed

**Database** (`internal/database/postgres.go`)
- Single connection pool initialization
- Shared across all repositories

**Config** (`internal/config/`)
- Environment variable loading
- Configuration struct with defaults

**Middleware** (`internal/middleware/`)
- `auth.go`: JWT token validation and extraction
- `validator.go`: Request body validation using go-playground/validator
- Applied globally in main.go via Fiber middleware chain

**Workers** (`internal/workers/`)
- `EmbeddingWorker`: Background worker that processes tasks without embeddings
- Runs every 5 minutes, batches up to 50 tasks at a time
- Auto-generates embeddings for tasks created without them

### Frontend Architecture

**Pages** (`src/pages/`)
- Top-level route components
- `TaskList.tsx`, `TaskDetail.tsx`, `Projects.tsx`, `Teams.tsx`, etc.
- Use React Router for navigation

**Components** (`src/components/`)
- Reusable UI components (buttons, forms, cards)
- Tailwind + lucide-react icons

**Services** (`src/services/`)
- `api.ts`: Axios client with typed endpoints
- Handles authentication header injection
- Error handling and response transformation

**Store** (`src/store/`)
- Zustand stores for state management
- `authStore.ts` for user/token state

**Types** (`src/types/`)
- TypeScript types and interfaces for API responses
- Mirrors backend models

**Testing**
- Unit tests: Vitest + React Testing Library (`.test.tsx` files)
- E2E tests: Playwright (in `e2e/` directory, targets `http://localhost:3000`)
- Setup file: `src/test/setup.ts`

## Key Conventions

### Backend (Go)

**Error Handling**
- Return errors from function calls, check explicitly
- Handler functions return `error` to Fiber for HTTP response handling
- Use `c.Status(statusCode).JSON(ErrorResponse{})` for API errors

**Dependency Injection**
- Services and repositories injected via constructor functions: `NewTaskService()`
- Main (`cmd/api/main.go`) wires all dependencies during startup
- Background workers started in main and stopped on graceful shutdown

**Swagger Documentation**
- API handlers annotated with Swagger comments (`@Summary`, `@Description`, `@Tags`, etc.)
- Generate docs: `make swagger` or `swag init -g cmd/api/main.go -o docs`
- Access at `http://localhost:8080/swagger/index.html` when API is running
- Update annotations in handler files to reflect in auto-generated docs

**Middleware**
- Auth context extracted via `getUserContext()` helper in handlers
- User ID and Organization ID passed down through service calls

**Database Queries**
- Use parameterized queries with `$1`, `$2` etc. to prevent SQL injection
- Context passed for timeout management: `r.db.QueryRow(ctx, query, args...)`

**Vector Operations**
- Task descriptions stored as pgvector embeddings via `GenerateEmbedding()` from embedding service
- Similarity search uses pgvector's distance operations in SQL queries
- Results ranked by cosine distance

### Frontend (React + TypeScript)

**React Query Patterns**
- All data fetching via `useQuery` with descriptive `queryKey` arrays
- Mutations via `useMutation` for state changes
- `queryFn` calls typed API methods from `services/api.ts`

**State Management**
- Zustand for persistent state (`authStore`)
- React component state for UI-only data (search filters, pagination)
- Query state handled by React Query (loading, error, data)

**Type Safety**
- All API responses have matching TypeScript types in `src/types/`
- Request bodies passed as typed objects to API methods
- Components use TypeScript props interfaces

**Styling**
- Tailwind CSS utility classes
- `clsx()` for conditional classNames
- Custom components in `src/components/`
- Icons from lucide-react

**API Calls**
- Centered in `src/services/api.ts`
- Use axios instance with interceptors for JWT auth
- Base URL: `http://localhost:8080/api/v1` (dev) or environment-based

**Component Organization**
- Functional components with hooks only
- Custom hooks can be added to `src/hooks/`
- Pages import from components, services, and types

### Embedding Service Integration

**When Creating/Updating Tasks**
- Backend calls embedding service via HTTP to `EMBEDDING_SERVICE_URL` (default: `http://localhost:8000`)
- Sends task description, receives vector embedding
- Stores embedding in PostgreSQL `description_embedding` column (pgvector type)

**Similarity Search**
- Used to find related tasks when creating new ones
- Query uses pgvector's `<->` distance operator
- Results ordered by distance, configurable limit

**Model Configuration**
- Model name controlled via `MODEL_NAME` environment variable in docker-compose
- Default: `all-MiniLM-L6-v2` (384 dimensions)
- Change via `.env` or docker-compose before running

## Database Schema & Migrations

### Schema
- **tasks**: Core entity with `description_embedding` (vector type) column
- **task_links**: Related tasks with `link_type` enum (similar, dependent, retry, related)
- **justifications**: Why a task is needed
- **blockers**: What blocks task progress, with `blocker_type` enum (approval, external_team, vendor, technical, resource, requirements)
- **reviews**: Task review records
- **acknowledgments**: User acknowledgment of tasks
- **projects**: Logical groupings
- **teams**: User teams with hierarchical roles (owner, manager, member, viewer)
- **team_members**: Junction table for team membership
- **project_teams**: Junction table for project-team assignments
- **users**: Authentication with password_hash, linked to organizations
- **organizations**: Multi-tenancy support

All tables include `created_at`, `updated_at`, and soft-delete support where applicable.

### Migrations
- Migration files in `backend/migrations/`
- `init.sql`: Base schema with extensions (uuid-ossp, pgvector), enums, and core tables
- `002_teams_and_projects.sql`: Teams and project-team associations
- `003_user_auth.sql`: Password authentication support
- Auto-applied on postgres container startup via docker-entrypoint-initdb.d
- **To reset**: `docker-compose down -v && docker-compose up -d postgres`

## Development Workflow

1. **Start services**: `make dev` (starts postgres, redis, embedding service)
2. **Apply migrations**: Auto-run on DB startup via `init.sql`
3. **Run API locally**: 
   - Standard: `make api` (compiles Go, starts on port 8080)
   - With hot reload: `cd backend && air` (auto-recompiles on file changes)
4. **Run frontend locally**: `make frontend` (Vite dev server on port 3000)
5. **Check API docs**: Visit `http://localhost:8080/swagger/index.html`
6. **Generate test JWT**: `curl -X POST http://localhost:8080/api/v1/auth/dev-token`
7. **Run tests**: `make test` or by service
8. **Check linting**: `make lint`
9. **Update Swagger docs**: `make swagger` (after adding/modifying handler annotations)

## Environment Variables

See `.env.example` for all options. Key variables:
- `SERVER_PORT`: Backend port (default: 8080)
- `ENVIRONMENT`: `development` or `production`
- `DB_*`: PostgreSQL connection details
- `JWT_SECRET`: Secret for token signing (change in production)
- `EMBEDDING_SERVICE_URL`: Embedding service endpoint
- `ALLOW_ORIGINS`: CORS allowed origins

## CI/CD

GitHub Actions workflows in `.github/workflows/`:
- **ci.yml**: Runs on push/PR to main/develop branches
  - `backend-test`: Go tests with PostgreSQL service container, uploads coverage to Codecov
  - `backend-lint`: golangci-lint with caching
  - `frontend-test`: npm tests (unit tests)
  - `frontend-lint`: ESLint and type checking
  - `frontend-e2e`: Playwright E2E tests (requires full stack running)
- **ci-cd.yaml**: Additional deployment pipeline (build Docker images, deploy to staging/production)

## Recommended MCP Servers

For enhanced Copilot capabilities with this codebase, consider configuring these MCP servers:

### PostgreSQL MCP Server
**Why**: Direct database querying, schema inspection, and vector operations
- Query tasks with embeddings: `SELECT * FROM tasks WHERE description_embedding IS NOT NULL`
- Inspect vector similarity: `SELECT * FROM tasks ORDER BY description_embedding <-> '[0.1, 0.2, ...]' LIMIT 5`
- Check migration status, analyze query performance
- **Connection**: `postgresql://goplan:goplan@localhost:5432/goplan` (dev environment)

### Playwright MCP Server
**Why**: E2E test authoring and debugging assistance
- Generate test code from UI interactions
- Debug failing tests with trace analysis
- **Config**: Tests in `frontend/e2e/`, runs against `http://localhost:3000`

### Docker MCP Server (if available)
**Why**: Container management and debugging
- Inspect running containers, view logs
- Check service health and resource usage
- **Services**: goplan-postgres, goplan-redis, goplan-embedding, goplan-api

## Important Notes

- Frontend uses **Yarn** (version 4.12.0) as package manager, not npm
- Go tests require `golangci-lint` installed: `brew install golangci-lint` (macOS) or use Docker for CI/CD
- Docker Compose setup includes health checks for all services
- API documentation auto-generated via Swagger/Swag; update handler comments to reflect in docs
- Embeddings generated synchronously during task creation, with fallback background worker processing missing embeddings every 5 minutes
- Air hot reload configured in `.air.toml` for faster backend development (excludes `_test.go` files)
- Migrations are applied once on container initialization; reset requires volume deletion
