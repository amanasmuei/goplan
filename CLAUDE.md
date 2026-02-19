# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoPlan is an AI-powered strategic planning platform that transforms vague ideas into structured, actionable strategy documents. Users describe a goal or business idea in plain text, and the system generates a comprehensive 5-section strategic plan using Claude AI.

**Stack:**
- **Backend** (`backend/`) -- Go + Fiber framework, PostgreSQL 16 (pgvector), Redis 7, Claude AI
- **Frontend** (`frontend-next/`) -- Next.js 16, TypeScript, Tailwind CSS, shadcn/ui
- **Embedding Service** (`embedding-service/`) -- Python FastAPI, sentence-transformers (all-MiniLM-L6-v2)

## Build Commands

### Backend (backend/)

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

### Frontend (frontend-next/)

```bash
cd frontend-next && yarn dev          # dev server (Next.js, port 3000)
cd frontend-next && yarn build        # production build
cd frontend-next && yarn test         # unit tests
cd frontend-next && yarn lint         # ESLint
cd frontend-next && yarn type-check   # TypeScript check
```

Package manager: **Yarn** (not npm) for the frontend.

### Database Migrations

```bash
# Using golang-migrate CLI
migrate -path ./backend/migrations -database "$DATABASE_URL" up
migrate -path ./backend/migrations -database "$DATABASE_URL" down 1

# Reset everything: docker-compose down -v && docker-compose up -d postgres
```

### Makefile (shortcuts)

```bash
make dev            # start postgres, redis, embedding-service via docker-compose
make api            # run Fiber backend locally
make frontend       # run Next.js frontend dev server
make build          # build backend binary + frontend production build
make test           # run backend + frontend tests
make lint           # run backend + frontend linters
make docker-up      # start all Docker services
make docker-down    # stop all Docker services
make swagger        # regenerate Swagger docs
make install        # download Go + Node dependencies
make migrate-up     # run database migrations
make migrate-down   # rollback one migration
```

## Architecture

### Backend (Fiber API)

```
backend/cmd/api/main.go          -> Entry point, Fiber app, route registration
backend/internal/handlers/       -> HTTP handlers (Swagger-annotated)
  auth_handler.go                   Auth (register, login, dev-token)
  strategy_handler.go               Strategy CRUD, generation, refinement, search
  subscription_handler.go           Subscription/billing management
  common.go                         Shared helpers (error responses, auth context)
backend/internal/services/       -> Business logic
  strategy_service.go               AI plan generation, refinement, similarity search
  prompt_builder.go                 3-layer prompt construction for Claude
  embedding_client.go               HTTP client for embedding service
backend/internal/claude/         -> Claude AI client
  client.go                         Anthropic API wrapper
  helpers.go                        Response parsing utilities
backend/internal/repository/     -> PostgreSQL data access (pgx/v5 direct queries)
  plan_repository.go                Strategic plans (with pgvector similarity)
  section_repository.go             Plan sections
  version_repository.go             Section + plan version history
  generation_log_repository.go      AI generation audit trail
  subscription_repository.go        User subscriptions
  user_repository.go                User accounts
backend/internal/models/         -> Domain structs and request/response types
  strategic_plan.go                 Plan entity, filters, responses
  plan_section.go                   Section types + content structure
  section_content.go                Typed content schemas per section
  version.go                        Section versions + plan snapshots
  subscription.go                   Subscription tiers and limits
  user.go                           User entity
backend/internal/middleware/     -> JWT auth, request validation
backend/internal/workers/        -> Background workers
  strategy_embedding_worker.go      Generates embeddings for plans without them
backend/internal/config/         -> Environment config loading
backend/internal/database/       -> Connection pool (pgxpool)
backend/internal/utils/          -> Shared utilities
```

Swagger docs available at `http://localhost:8080/swagger/index.html` when backend is running.

### Strategy Generation Flow

1. User submits plain-text input describing their goal
2. `PromptBuilder` constructs a 3-layer prompt:
   - **System prompt**: Expert strategy consultant persona + JSON output schema
   - **Category overlay**: Domain-specific guidance (business, SaaS, event, etc.)
   - **Depth instruction**: Standard or deep analysis mode
3. Claude generates a structured JSON response with classification + 5 sections
4. `StrategyService` parses the response, creates the plan, and stores sections
5. Each section is versioned; users can refine individual sections with follow-up prompts
6. Background `StrategyEmbeddingWorker` generates pgvector embeddings for similarity search

### Section Types (5 per plan)

| Order | Type | Description |
|-------|------|-------------|
| 1 | `executive_brief` | High-level summary and key takeaways |
| 2 | `strategic_context` | Market analysis, competitive landscape |
| 3 | `recommended_approach` | Core strategy and alternatives considered |
| 4 | `phased_execution` | Timeline with phases, milestones, KPIs |
| 5 | `immediate_action` | First 30-day action items |

### Frontend (Next.js App Router)

```
frontend-next/src/app/
  (auth)/                        -> Auth layout (login, register)
  (dashboard)/                   -> Authenticated layout with sidebar
    dashboard/page.tsx              Dashboard / plan list
    strategies/new/page.tsx         New strategy creation
    strategies/[id]/page.tsx        Strategy viewer (3-panel section display)
    strategies/[id]/versions/       Version history page
  layout.tsx                     -> Root layout
  page.tsx                       -> Landing / redirect

frontend-next/src/components/
  auth/AuthGuard.tsx             -> Route protection
  layout/Sidebar.tsx             -> Navigation sidebar
  strategy/                      -> Strategy-specific components
  ui/                            -> shadcn/ui primitives

frontend-next/src/hooks/         -> React Query hooks (useStrategy, useSubscription)
frontend-next/src/lib/           -> Utilities, types, API client
frontend-next/src/store/         -> Zustand stores (auth)
frontend-next/src/providers/     -> Context providers
```

### Data Model

Core tables (PostgreSQL with pgvector):

- **strategic_plans** -- User plans with category, status, content_embedding (pgvector)
- **plan_sections** -- 5 sections per plan, each with structured JSON content
- **section_versions** -- Version history for individual sections (refinements)
- **plan_versions** -- Full plan snapshots at version boundaries
- **generation_logs** -- Audit trail for all AI generation requests
- **subscriptions** -- User subscription tiers (free, pro, pro_plus)
- **users** -- User accounts with auth

Plan categories: `business`, `saas`, `event`, `nonprofit`, `personal`, `education`, `real_estate`, `generic`

Plan statuses: `draft` -> `generating` -> `complete` -> `archived`

## Design Constraints

- **Structured over conversational**: Generate complete structured documents, not chat responses
- **Framework integrity**: Each plan always has exactly 5 sections in defined order
- **Version everything**: Every section refinement creates a new version; plans are snapshotted
- **Full auditability**: Every AI generation logged to `generation_logs` with token usage
- **Parameterized queries only** -- use `$1`, `$2` etc. for all SQL; no string interpolation
- **Subscription-gated features**: Plan creation limits and deep analysis tied to subscription tier

## Key Conventions

- Repository pattern: concrete structs in `backend/internal/repository/` using `pgxpool.Pool`
- Fiber handlers return `error`; use `c.Status(code).JSON(ErrorResponse{})` for errors
- Auth context extracted via `getUserContext()` helper in handlers
- Strategy service orchestrates Claude calls, parsing, storage, and versioning
- Prompt builder centralizes all AI prompt construction
- Section types are strongly typed constants in `models/plan_section.go`
- Embedding worker runs in background goroutine on configurable interval
- Config loaded from environment variables via `internal/config/config.go`
