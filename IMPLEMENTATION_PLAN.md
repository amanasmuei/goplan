# GoPlan Implementation Plan

> **Plan anything. Track everything.**

This document provides a comprehensive implementation roadmap for building GoPlan - a universal, AI-powered project management platform with MCP (Model Context Protocol) architecture.

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Architecture Overview](#2-architecture-overview)
3. [Technology Stack](#3-technology-stack)
4. [Project Structure](#4-project-structure)
5. [Implementation Phases](#5-implementation-phases)
6. [Detailed Task Breakdown](#6-detailed-task-breakdown)
7. [User Journeys](#7-user-journeys)
8. [Data Model](#8-data-model)
9. [API Specification](#9-api-specification)
10. [AI Integration](#10-ai-integration)
11. [Security Architecture](#11-security-architecture)
12. [Observability](#12-observability)
13. [Testing Strategy](#13-testing-strategy)
14. [Deployment Strategy](#14-deployment-strategy)
15. [Success Metrics](#15-success-metrics)

---

## Current Implementation Status

> **Last Updated**: January 28, 2026

### Progress Overview

| Phase | Status | Progress |
|-------|--------|----------|
| Phase 0: Foundation | ✅ Complete | 100% |
| Phase 1: Core Domain | ✅ Complete | 100% |
| Phase 2: MCP Server | ✅ Complete | 100% |
| Phase 3: Persistence | ✅ Complete | 100% |
| Phase 4: API Layer | ✅ Complete | 100% |
| Phase 5: AI Integration | ✅ Complete | 100% |
| Phase 6: Frontend MVP | ✅ Complete | 100% |
| Phase 7: Production | ✅ Complete | 100% |

### Completed Components

**Backend (Go)**
- ✅ Project structure with `go.mod`
- ✅ Domain entities (User, Workspace, Plan, Task)
- ✅ MCP types and ToolRegistry
- ✅ Configuration management (Viper + godotenv)
- ✅ HTTP server entry point
- ✅ Repository layer (sqlc + pgx/v5)
- ✅ MCP tools (17 tools: workspace, plan, task, milestone, activity)
- ✅ REST API handlers (28 endpoints)
- ✅ Auth middleware (header-based + RBAC)
- ✅ Claude API client with streaming
- ✅ MCP-to-Claude bridge (tool execution)
- ✅ Human-in-the-loop safety layer
- ✅ AI service (chat, plan generation, task breakdown)
- ✅ AI REST endpoints (8 endpoints)

**Frontend (Next.js 16)**
- ✅ Next.js 16 project with App Router
- ✅ TypeScript types matching contracts
- ✅ Type-safe API client
- ✅ shadcn/ui components
- ✅ Auth pages (login/signup) - Wired to `/api/v1/auth/*` endpoints
- ✅ Auth hook (`useAuth`) with login/signup/logout functionality
- ✅ Dashboard layout
- ✅ React Query hooks (workspaces, plans, tasks, comments)
- ✅ Kanban view with drag-and-drop
- ✅ List view with sorting/filtering
- ✅ Calendar view
- ✅ Task components (card, row, detail, form)
- ✅ Workspace management pages

**Database & DevOps**
- ✅ 7 SQL migrations (up/down)
- ✅ Docker Compose (PostgreSQL, Redis, Mailhog)
- ✅ Dockerfile (multi-stage, optimized)
- ✅ Makefile with dev commands
- ✅ GitHub Actions CI/CD pipeline
- ✅ Kubernetes manifests (Deployment, Service, Ingress, HPA)
- ✅ Helm chart with multi-environment support
- ✅ Deployment scripts (deploy.sh, rollback.sh, health-check.sh)

**Observability & Security**
- ✅ Structured logging (log/slog with JSON)
- ✅ Prometheus metrics (/metrics endpoint)
- ✅ Health checks (/health, /ready, /healthz)
- ✅ JWT authentication (access + refresh tokens)
  - `/api/v1/auth/signup` - User registration with bcrypt password hashing
  - `/api/v1/auth/login` - User login with JWT token pair generation
  - `/api/v1/auth/refresh` - Token refresh endpoint
  - `/api/v1/auth/logout` - Logout (client-side token discard)
- ✅ Rate limiting (per-IP and per-user)
- ✅ Input validation & sanitization
- ✅ Security headers (CSP, HSTS, X-Frame-Options)
- ✅ Audit logging
- ✅ Graceful shutdown

### Next Steps

1. ~~**Repository Layer** - Connect domain entities to PostgreSQL~~ ✅
2. ~~**REST API Handlers** - Implement endpoints from contracts~~ ✅
3. ~~**Auth Middleware** - JWT validation and RBAC~~ ✅
4. ~~**Claude Integration** - Connect MCP server to Claude API~~ ✅
5. ~~**Frontend Integration** - Wire frontend to REST API~~ ✅
6. ~~**Phase 7: Production** - Deployment, monitoring, optimization~~ ✅
7. ~~**JWT Authentication** - Replace header-based auth with proper JWT~~ ✅

**🎉 ALL PHASES COMPLETE - GoPlan MVP is ready for deployment!**

### Post-MVP Enhancements (Optional)
- End-to-end testing across all layers
- Performance optimization & load testing
- Additional OAuth providers (Google, GitHub)
- Real-time updates with WebSocket
- Mobile app (React Native)

---

## 1. Executive Summary

### Vision
GoPlan is a flexible, modular planning platform that adapts to how people actually work, supporting software projects, events, operations, marketing campaigns, and more.

### Core Differentiators
- **Universal**: One platform for all project types
- **AI-Native**: MCP-first architecture with human-in-the-loop safety
- **Flexible**: Configuration over enforced methodology
- **Enterprise-Ready**: RBAC, audit logs, deterministic AI behavior

### MVP Scope
- Plan & Task CRUD with customizable workflows
- 4 views: Kanban, List, Calendar, Gantt
- AI-assisted plan creation and task breakdown
- Real-time collaboration
- Human-in-the-loop AI safety

---

## 2. Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLIENTS                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                │
│  │   Web    │  │  Mobile  │  │   CLI    │  │   Slack  │                │
│  │ (Next.js)│  │  (React  │  │  (MCP)   │  │   Bot    │                │
│  │          │  │  Native) │  │          │  │          │                │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘                │
└───────┼─────────────┼─────────────┼─────────────┼───────────────────────┘
        │             │             │             │
        └─────────────┴──────┬──────┴─────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────────────┐
│                      API GATEWAY                                         │
│  ┌─────────────────────────┴─────────────────────────────────────────┐  │
│  │  • Authentication (JWT/OAuth2)  • Rate Limiting                    │  │
│  │  • Request Validation           • CORS/Security Headers            │  │
│  │  • Request Routing              • Metrics Collection               │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│   REST API    │   │  WebSocket    │   │  MCP Server   │
│   Handlers    │   │   Gateway     │   │  (AI Agents)  │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
┌───────────────────────────┼───────────────────────────────────────────┐
│                    APPLICATION LAYER                                   │
│  ┌────────────────┐ ┌────────────────┐ ┌────────────────┐             │
│  │  Plan Service  │ │  Task Service  │ │  User Service  │             │
│  └────────────────┘ └────────────────┘ └────────────────┘             │
│  ┌────────────────┐ ┌────────────────┐ ┌────────────────┐             │
│  │ Intent Service │ │ Search Service │ │Notification Svc│             │
│  └────────────────┘ └────────────────┘ └────────────────┘             │
└───────────────────────────┬───────────────────────────────────────────┘
                            │
┌───────────────────────────┼───────────────────────────────────────────┐
│                      DOMAIN LAYER                                      │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │  Entities: Workspace, Plan, Phase, Task, Milestone, User        │  │
│  │  Value Objects: Status, Priority, TaskType, Permission          │  │
│  │  Domain Events: PlanCreated, TaskAssigned, StatusChanged        │  │
│  │  Domain Services: WorkflowEngine, PermissionChecker             │  │
│  └─────────────────────────────────────────────────────────────────┘  │
└───────────────────────────┬───────────────────────────────────────────┘
                            │
┌───────────────────────────┼───────────────────────────────────────────┐
│                  INFRASTRUCTURE LAYER                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │  PostgreSQL  │  │    Redis     │  │   Claude     │                 │
│  │  Repository  │  │    Cache     │  │   API        │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │    Event     │  │   Email      │  │   Storage    │                 │
│  │    Bus       │  │   Service    │  │   (S3)       │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
└───────────────────────────────────────────────────────────────────────┘
```

### MCP Multi-Agent Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         USER INPUT                                   │
│              "Plan a product launch for Q2"                         │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       INTENT AGENT                                   │
│  • Parse natural language                                            │
│  • Extract entities (plan_name, plan_type, timeline)                │
│  • Calculate confidence score                                        │
│  • Generate MCPIntentEnvelope                                        │
└─────────────────────────────┬───────────────────────────────────────┘
                              │ confidence: 0.94
                              │ intent_type: CREATE_PLAN
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       PLANNER AGENT                                  │
│  • Decompose high-level goals                                        │
│  • Generate phases and milestones                                    │
│  • Create task breakdown                                             │
│  • Suggest dependencies                                              │
└─────────────────────────────┬───────────────────────────────────────┘
                              │ proposed_actions: [task.create, ...]
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      EXECUTOR AGENT                                  │
│  • Validate permissions                                              │
│  • Stage tool calls                                                  │
│  • Create DRAFT entities                                             │
│  • Await user approval                                               │
└─────────────────────────────┬───────────────────────────────────────┘
                              │ status: DRAFT
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    USER APPROVAL                                     │
│  [✓ Approve]  [✎ Edit]  [✗ Reject]                                  │
└─────────────────────────────┬───────────────────────────────────────┘
                              │ approved: true
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      MCP TOOL EXECUTION                              │
│  • Execute approved actions                                          │
│  • Record in audit log                                               │
│  • Emit domain events                                                │
│  • Notify stakeholders                                               │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 3. Technology Stack

### Backend (Go)

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | Go 1.22+ | Performance, concurrency, simplicity |
| HTTP Router | Chi | Lightweight, stdlib compatible |
| Database Access | sqlc | Type-safe SQL, no ORM overhead |
| Migrations | golang-migrate | Version-controlled schema changes |
| Validation | go-playground/validator | Struct tag validation |
| Logging | zap | High-performance structured logging |
| Config | Viper | 12-factor config management |
| Testing | testify + gomock | Assertions and mocking |
| Observability | OpenTelemetry | Tracing, metrics, logs |

### Database & Storage

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Primary DB | PostgreSQL 16 | JSONB, full-text search, reliability |
| Cache | Redis 7 | Sessions, pub/sub, rate limiting |
| Search | PostgreSQL FTS (MVP) / Elasticsearch (v2) | Scalable search |
| File Storage | S3-compatible | Attachments, exports |
| Message Queue | Redis Streams (MVP) / NATS (v2) | Event distribution |

### AI/ML

| Component | Technology | Rationale |
|-----------|------------|-----------|
| LLM | Claude API (Anthropic) | Best reasoning, safety |
| Embeddings | Voyage AI / OpenAI | Semantic search |
| Vector Store | pgvector (MVP) | Similar plan retrieval |
| Prompt Management | Custom (versioned) | Controlled prompts |

### Frontend

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Framework | Next.js 16 (App Router) | SSR, RSC, Turbopack stable, PPR, React 19 |
| React | React 19 | Actions, use hook, useOptimistic |
| Language | TypeScript 5.x | Type safety, satisfies operator |
| State | Zustand 5 | Simple, performant |
| Data Fetching | TanStack Query v5 | Caching, mutations, suspense |
| Forms | React Hook Form + Zod | Validation, type inference |
| Styling | Tailwind CSS v4 + shadcn/ui | Rapid development |
| Kanban | dnd-kit | Drag-and-drop |
| Calendar | FullCalendar | Feature-rich |
| Gantt | Frappe Gantt | Lightweight |
| Real-time | Socket.io / native WebSocket | Live updates |

### Infrastructure

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Containers | Docker | Consistency |
| Orchestration | Kubernetes | Scalability |
| CI/CD | GitHub Actions | Integration |
| IaC | Terraform | Reproducible |
| Secrets | HashiCorp Vault / AWS SM | Security |
| CDN | Cloudflare | Performance |

---

## 4. Project Structure

```
goplan/
├── cmd/
│   ├── server/              # Main API server entry point
│   │   └── main.go
│   ├── worker/              # Background job processor
│   │   └── main.go
│   └── migrate/             # Database migration CLI
│       └── main.go
│
├── internal/
│   ├── domain/              # Core business logic (no dependencies)
│   │   ├── workspace/
│   │   │   ├── workspace.go
│   │   │   ├── repository.go
│   │   │   └── events.go
│   │   ├── plan/
│   │   │   ├── plan.go
│   │   │   ├── phase.go
│   │   │   ├── repository.go
│   │   │   └── events.go
│   │   ├── task/
│   │   │   ├── task.go
│   │   │   ├── subtask.go
│   │   │   ├── repository.go
│   │   │   └── events.go
│   │   ├── user/
│   │   │   ├── user.go
│   │   │   ├── permission.go
│   │   │   └── repository.go
│   │   └── shared/
│   │       ├── status.go
│   │       ├── priority.go
│   │       └── errors.go
│   │
│   ├── application/         # Use cases / Application services
│   │   ├── plan/
│   │   │   ├── create_plan.go
│   │   │   ├── update_plan.go
│   │   │   ├── list_plans.go
│   │   │   └── dto.go
│   │   ├── task/
│   │   │   ├── create_task.go
│   │   │   ├── assign_task.go
│   │   │   └── dto.go
│   │   └── intent/
│   │       ├── process_intent.go
│   │       └── dto.go
│   │
│   ├── infrastructure/      # External adapters
│   │   ├── postgres/
│   │   │   ├── plan_repo.go
│   │   │   ├── task_repo.go
│   │   │   ├── queries/     # sqlc queries
│   │   │   └── migrations/
│   │   ├── redis/
│   │   │   ├── cache.go
│   │   │   └── pubsub.go
│   │   ├── llm/
│   │   │   ├── claude.go
│   │   │   └── prompts/
│   │   ├── email/
│   │   │   └── sendgrid.go
│   │   └── storage/
│   │       └── s3.go
│   │
│   ├── api/                 # HTTP layer
│   │   ├── rest/
│   │   │   ├── router.go
│   │   │   ├── middleware/
│   │   │   ├── handlers/
│   │   │   │   ├── plan_handler.go
│   │   │   │   ├── task_handler.go
│   │   │   │   └── auth_handler.go
│   │   │   └── dto/
│   │   └── websocket/
│   │       ├── hub.go
│   │       └── client.go
│   │
│   ├── mcp/                 # MCP Server implementation
│   │   ├── server.go
│   │   ├── registry.go      # Tool registry
│   │   ├── agents/
│   │   │   ├── intent_agent.go
│   │   │   ├── planner_agent.go
│   │   │   ├── executor_agent.go
│   │   │   └── analyst_agent.go
│   │   ├── tools/
│   │   │   ├── project_tools.go
│   │   │   ├── task_tools.go
│   │   │   └── plan_tools.go
│   │   └── audit/
│   │       └── logger.go
│   │
│   └── config/              # Configuration
│       └── config.go
│
├── pkg/                     # Public reusable packages
│   ├── mcp/                 # MCP types (can be imported)
│   │   ├── envelope.go
│   │   ├── action.go
│   │   └── context.go
│   └── validation/
│       └── validator.go
│
├── web/                     # Frontend (Next.js)
│   ├── src/
│   │   ├── app/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── lib/
│   │   └── types/
│   ├── package.json
│   └── tsconfig.json
│
├── migrations/              # SQL migrations
│   ├── 000001_init.up.sql
│   ├── 000001_init.down.sql
│   └── ...
│
├── docs/
│   ├── api/                 # OpenAPI specs
│   │   └── openapi.yaml
│   ├── adr/                 # Architecture Decision Records
│   └── diagrams/
│
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── kubernetes/
│       ├── base/
│       └── overlays/
│
├── scripts/
│   ├── dev-setup.sh
│   ├── generate-sqlc.sh
│   └── seed-db.sh
│
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.md
└── README.md
```

---

## 5. Implementation Phases

### Phase 0: Foundation (Infrastructure Setup) ✅ COMPLETE
**Duration: Sprint 1**

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 0: FOUNDATION                                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Project    │  │  Dev Env     │  │   CI/CD      │          │
│  │   Setup      │  │  (Docker)    │  │   Pipeline   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Database   │  │  Config      │  │ Observability│          │
│  │   Schema     │  │  Management  │  │   Setup      │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 1: Core Domain Layer 🔄 IN PROGRESS
**Duration: Sprint 2**

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 1: DOMAIN LAYER                                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Entities           Value Objects        Domain Services         │
│  ┌────────────┐    ┌────────────┐       ┌────────────┐          │
│  │ Workspace  │    │   Status   │       │  Workflow  │          │
│  │ Plan       │    │  Priority  │       │   Engine   │          │
│  │ Phase      │    │ Permission │       │            │          │
│  │ Task       │    │  TaskType  │       │ Permission │          │
│  │ Milestone  │    └────────────┘       │  Checker   │          │
│  │ User       │                         └────────────┘          │
│  └────────────┘                                                  │
│                                                                  │
│  Domain Events                  Repository Interfaces            │
│  ┌────────────────────────┐    ┌────────────────────────┐       │
│  │ PlanCreated            │    │ PlanRepository         │       │
│  │ TaskAssigned           │    │ TaskRepository         │       │
│  │ StatusChanged          │    │ UserRepository         │       │
│  │ MilestoneReached       │    │ WorkspaceRepository    │       │
│  └────────────────────────┘    └────────────────────────┘       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 2: MCP Server Core 🔄 IN PROGRESS
**Duration: Sprint 3-4**

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 2: MCP SERVER                                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Tool Registry              Core Tools                           │
│  ┌──────────────────┐      ┌──────────────────┐                 │
│  │ Register()       │      │ project.create   │                 │
│  │ Get()            │      │ project.update   │                 │
│  │ ExecuteTool()    │      │ project.get      │                 │
│  │ ListTools()      │      │ project.list     │                 │
│  └──────────────────┘      │ task.create      │                 │
│                            │ task.update      │                 │
│  Intent Processing         │ task.move        │                 │
│  ┌──────────────────┐      │ task.list        │                 │
│  │ ValidateIntent() │      │ plan.generate    │                 │
│  │ RouteIntent()    │      └──────────────────┘                 │
│  │ ExecuteAction()  │                                           │
│  └──────────────────┘      Audit System                         │
│                            ┌──────────────────┐                 │
│                            │ RecordAudit()    │                 │
│                            │ QueryAuditLog()  │                 │
│                            └──────────────────┘                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 3: Persistence Layer 🔄 IN PROGRESS
**Duration: Sprint 4-5**

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 3: PERSISTENCE                                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  PostgreSQL          Redis Cache         Search                  │
│  ┌──────────────┐   ┌──────────────┐    ┌──────────────┐        │
│  │ Migrations   │   │ Session      │    │ Full-text    │        │
│  │ Repositories │   │ Cache        │    │ Indexing     │        │
│  │ Transactions │   │ Pub/Sub      │    │              │        │
│  │ Query Optim  │   │ Rate Limit   │    │              │        │
│  └──────────────┘   └──────────────┘    └──────────────┘        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 4: API Layer ⏳ PENDING
**Duration: Sprint 5-6**

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 4: API LAYER                                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  REST API             Authentication        WebSocket            │
│  ┌──────────────┐    ┌──────────────┐      ┌──────────────┐     │
│  │ OpenAPI Spec │    │ JWT/OAuth2   │      │ Hub          │     │
│  │ Handlers     │    │ RBAC         │      │ Rooms        │     │
│  │ Validation   │    │ Session Mgmt │      │ Broadcast    │     │
│  │ Pagination   │    │ Rate Limit   │      │ Presence     │     │
│  └──────────────┘    └──────────────┘      └──────────────┘     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 5: AI Integration ⏳ PENDING
**Duration: Sprint 6-7**

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 5: AI INTEGRATION                                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Intent Agent           Planner Agent        Executor Agent      │
│  ┌──────────────┐      ┌──────────────┐     ┌──────────────┐    │
│  │ NL Parsing   │      │ Goal Decomp  │     │ Permission   │    │
│  │ Entity Extrc │      │ Task Breakdn │     │ Tool Staging │    │
│  │ Confidence   │      │ Suggestions  │     │ Draft Create │    │
│  └──────────────┘      └──────────────┘     └──────────────┘    │
│                                                                  │
│  Prompt Management      Human-in-Loop                            │
│  ┌──────────────┐      ┌──────────────┐                         │
│  │ Templates    │      │ Draft Review │                         │
│  │ Versioning   │      │ Approval UI  │                         │
│  │ Context Bld  │      │ Rejection    │                         │
│  └──────────────┘      └──────────────┘                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 6: Frontend MVP 🔄 IN PROGRESS
**Duration: Sprint 7-10**

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 6: FRONTEND                                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Core UI                Views                                    │
│  ┌──────────────┐      ┌──────────────────────────────────┐     │
│  │ Auth Flow    │      │  ┌────────┐ ┌────────┐           │     │
│  │ Navigation   │      │  │ Kanban │ │  List  │           │     │
│  │ Plan CRUD    │      │  └────────┘ └────────┘           │     │
│  │ Task CRUD    │      │  ┌────────┐ ┌────────┐           │     │
│  │ Settings     │      │  │Calendar│ │ Gantt  │           │     │
│  └──────────────┘      │  └────────┘ └────────┘           │     │
│                        └──────────────────────────────────┘     │
│                                                                  │
│  AI Chat                Collaboration                            │
│  ┌──────────────┐      ┌──────────────┐                         │
│  │ Chat Panel   │      │ Real-time    │                         │
│  │ Draft Review │      │ Comments     │                         │
│  │ Suggestions  │      │ Activity     │                         │
│  └──────────────┘      └──────────────┘                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 7: Production Readiness ⏳ PENDING
**Duration: Sprint 11-12**

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 7: PRODUCTION                                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Security              Performance          Operations           │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐     │
│  │ Pen Testing  │     │ Load Testing │     │ Monitoring   │     │
│  │ OWASP Audit  │     │ Optimization │     │ Alerting     │     │
│  │ SOC2 Prep    │     │ CDN Setup    │     │ Runbooks     │     │
│  │ Encryption   │     │ DB Tuning    │     │ Backup/DR    │     │
│  └──────────────┘     └──────────────┘     └──────────────┘     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Task Breakdown

### Phase 0: Foundation ✅ COMPLETE

#### 0.1 Project Setup
- [x] Initialize Go module (`go mod init github.com/goplan/goplan`)
- [x] Create directory structure per Section 4
- [x] Setup Makefile with common targets
- [x] Configure .gitignore, .editorconfig
- [ ] Setup pre-commit hooks (golangci-lint, gofmt)

#### 0.2 Development Environment
- [x] Create Docker Compose for local development
  - PostgreSQL 16
  - Redis 7
  - Mailhog (email testing)
- [ ] Create dev-setup.sh script
- [ ] Document setup in README.md

#### 0.3 CI/CD Pipeline
- [x] GitHub Actions workflow for:
  - Lint (golangci-lint)
  - Test (with coverage)
  - Build (multi-arch)
  - Security scan (trivy, gosec)
- [ ] Setup branch protection rules

#### 0.4 Database Foundation
- [x] Install golang-migrate
- [x] Create initial migration with core tables
- [ ] Setup sqlc with configuration
- [ ] Create seed data script

#### 0.5 Configuration Management
- [x] Setup Viper for config loading
- [x] Create config struct with validation
- [x] Support env vars, config files
- [x] Document all config options (in contracts/03-environment-config.md)

#### 0.6 Observability Foundation
- [ ] Integrate zap logger with structured output
- [ ] Add request ID middleware
- [ ] Setup OpenTelemetry tracer
- [ ] Create health check endpoint

---

### Phase 1: Core Domain Layer 🔄 IN PROGRESS (70%)

#### 1.1 Workspace Entity
- [x] Define Workspace struct with fields:
  - ID (UUID)
  - Name, Slug
  - OwnerID
  - Settings (JSONB)
  - CreatedAt, UpdatedAt
- [ ] Define WorkspaceRepository interface
- [x] Add validation rules
- [ ] Define WorkspaceCreated, WorkspaceUpdated events

#### 1.2 Plan Entity
- [x] Define Plan struct with fields:
  - ID, WorkspaceID
  - Name, Description
  - Domain (software/event/ops/collection/generic)
  - Status (draft/active/on_hold/completed/archived)
  - OwnerID
  - StartDate, EndDate
  - CustomStatuses []Status
  - CustomFields []FieldDefinition
  - Tags []string
  - CreatedAt, UpdatedAt
- [ ] Define PlanRepository interface
- [ ] Add Status state machine
- [ ] Define Plan domain events

#### 1.3 Phase Entity
- [ ] Define Phase struct:
  - ID, PlanID
  - Name, Description
  - Order
  - StartDate, EndDate
- [ ] Define PhaseRepository interface

#### 1.4 Task Entity
- [x] Define Task struct:
  - ID, PlanID, PhaseID (optional)
  - ParentTaskID (for subtasks)
  - Title, Description
  - Status (custom per plan)
  - Priority (low/medium/high/critical)
  - AssigneeID
  - DueDate
  - EstimatedHours
  - Tags []string
  - CustomFieldValues map[string]interface{}
  - Dependencies []TaskID
  - CreatedAt, UpdatedAt
- [ ] Define TaskRepository interface
- [ ] Add dependency cycle detection
- [ ] Define Task domain events

#### 1.5 Milestone Entity
- [ ] Define Milestone struct:
  - ID, PlanID
  - Name, Description
  - DueDate
  - Status
  - LinkedTaskIDs []TaskID
- [ ] Define MilestoneRepository interface

#### 1.6 User & Permission
- [x] Define User struct:
  - ID, Email, Name
  - PasswordHash
  - AvatarURL
  - CreatedAt, UpdatedAt
- [x] Define Permission value objects
- [x] Define workspace membership (user, workspace, role)
- [ ] Define plan-level permissions

#### 1.7 Domain Services
- [ ] Implement WorkflowEngine for status transitions
- [ ] Implement PermissionChecker service
- [ ] Implement DependencyValidator

#### 1.8 Domain Events
- [ ] Create event interfaces
- [ ] Implement in-memory event bus
- [ ] Add event handlers for audit logging

---

### Phase 2: MCP Server Core 🔄 IN PROGRESS (40%)

#### 2.1 Complete MCP Types
- [x] Move types to pkg/mcp for reusability
- [ ] Add JSON schema validation
- [ ] Add OpenAPI annotations

#### 2.2 Tool Registry
```go
type ToolRegistry struct {
    tools map[string]MCPTool
    mu    sync.RWMutex
}

func (r *ToolRegistry) Register(tool MCPTool) error
func (r *ToolRegistry) Get(name string) (MCPTool, error)
func (r *ToolRegistry) ExecuteTool(name string, ctx MCPExecutionContext, args map[string]interface{}) (interface{}, error)
func (r *ToolRegistry) ListTools() []ToolDescriptor
```
- [x] Implement ToolRegistry struct
- [ ] Add tool validation
- [ ] Add execution timeout
- [ ] Add metrics collection

#### 2.3 Project Tools
- [ ] Implement `project.create` tool
- [ ] Implement `project.update` tool
- [ ] Implement `project.get` tool
- [ ] Implement `project.list` tool
- [ ] Implement `project.archive` tool

#### 2.4 Task Tools
- [ ] Implement `task.create` tool
- [ ] Implement `task.update` tool
- [ ] Implement `task.move` tool (status change)
- [ ] Implement `task.get` tool
- [ ] Implement `task.list` tool
- [ ] Implement `task.assign` tool
- [ ] Implement `task.add_dependency` tool

#### 2.5 Plan AI Tools
- [ ] Implement `plan.generate_tasks` tool
- [ ] Implement `plan.suggest_tasks` tool
- [ ] Implement `plan.analyze_risks` (stub for MVP)

#### 2.6 Intent Validation & Routing
- [x] Enhance ValidateIntent with JSON schema
- [ ] Add entity validation per intent type
- [ ] Implement confidence thresholds
- [ ] Add clarification question generation

#### 2.7 Audit System
- [x] Create AuditLog table migration
- [ ] Implement persistent AuditLogger
- [ ] Add query methods (by user, workspace, time range)
- [ ] Add audit report generation

---

### Phase 3: Persistence Layer 🔄 IN PROGRESS (30%)

#### 3.1 PostgreSQL Migrations ✅ COMPLETE
```sql
-- 000001_init.up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- for search

CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    owner_id UUID NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    domain VARCHAR(50) NOT NULL DEFAULT 'generic',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    owner_id UUID NOT NULL,
    start_date DATE,
    end_date DATE,
    custom_statuses JSONB DEFAULT '[]',
    custom_fields JSONB DEFAULT '[]',
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ... more tables
```
- [ ] Create all table migrations
- [ ] Add indexes for common queries
- [ ] Add triggers for updated_at

#### 3.2 sqlc Setup
- [ ] Configure sqlc.yaml
- [ ] Write SQL queries for all entities
- [ ] Generate Go code
- [ ] Create repository implementations

#### 3.3 Repository Implementations
- [ ] Implement WorkspaceRepository
- [ ] Implement PlanRepository with filtering
- [ ] Implement TaskRepository with search
- [ ] Implement UserRepository
- [ ] Implement AuditLogRepository

#### 3.4 Redis Integration
- [ ] Setup Redis client with connection pooling
- [ ] Implement session store
- [ ] Implement cache layer for plans/tasks
- [ ] Implement pub/sub for real-time events
- [ ] Implement rate limiter

#### 3.5 Search
- [ ] Add PostgreSQL full-text search indexes
- [ ] Implement search query builder
- [ ] Add search highlighting

---

### Phase 4: API Layer ⏳ PENDING

#### 4.1 OpenAPI Specification
- [x] Define OpenAPI 3.1 spec for all endpoints (in contracts/02-api-endpoints.md)
- [x] Add request/response schemas
- [x] Document authentication
- [ ] Generate API docs

#### 4.2 REST API Handlers
- [ ] Auth endpoints:
  - POST /api/auth/signup
  - POST /api/auth/login
  - POST /api/auth/logout
  - POST /api/auth/refresh
  - POST /api/auth/forgot-password
- [ ] Workspace endpoints:
  - POST /api/workspaces
  - GET /api/workspaces
  - GET /api/workspaces/:id
  - PATCH /api/workspaces/:id
  - POST /api/workspaces/:id/invite
- [ ] Plan endpoints:
  - POST /api/workspaces/:wid/plans
  - GET /api/workspaces/:wid/plans
  - GET /api/plans/:id
  - PATCH /api/plans/:id
  - DELETE /api/plans/:id
- [ ] Task endpoints:
  - POST /api/plans/:pid/tasks
  - GET /api/plans/:pid/tasks
  - GET /api/tasks/:id
  - PATCH /api/tasks/:id
  - DELETE /api/tasks/:id
  - POST /api/tasks/:id/comments
- [ ] MCP endpoints:
  - POST /api/mcp/intent
  - POST /api/mcp/tools/execute
  - GET /api/mcp/tools

#### 4.3 Authentication
- [ ] Implement JWT token generation
- [ ] Add refresh token rotation
- [ ] Implement OAuth2 providers (Google, GitHub)
- [ ] Add session management
- [ ] Implement password reset flow

#### 4.4 Authorization (RBAC)
- [ ] Define roles: owner, admin, member, viewer
- [ ] Define permissions per resource
- [ ] Implement permission middleware
- [ ] Add plan-level permission overrides

#### 4.5 Middleware
- [ ] Request logging middleware
- [ ] Error handling middleware
- [ ] Rate limiting middleware
- [ ] CORS middleware
- [ ] Request validation middleware
- [ ] Pagination middleware

#### 4.6 WebSocket Gateway
- [ ] Implement connection hub
- [ ] Add room-based subscriptions (plan rooms)
- [ ] Implement presence tracking
- [ ] Add reconnection handling
- [ ] Implement event broadcasting

---

### Phase 5: AI Integration ⏳ PENDING

#### 5.1 LLM Client
- [ ] Create Claude API client wrapper
- [ ] Add retry logic with exponential backoff
- [ ] Implement streaming responses
- [ ] Add token counting
- [ ] Add cost tracking

#### 5.2 Prompt Templates
- [ ] Create system prompt for GoPlan AI
- [ ] Create CREATE_PLAN prompt template
- [ ] Create ADD_TASK prompt template
- [ ] Create SUGGEST_TASKS prompt template
- [ ] Create ASK_SUMMARY prompt template
- [ ] Implement prompt versioning

#### 5.3 Context Builder
- [ ] Build plan context from database
- [ ] Build task context with relationships
- [ ] Build user context with permissions
- [ ] Implement context truncation for token limits

#### 5.4 Intent Agent
- [ ] Implement natural language parsing
- [ ] Entity extraction (dates, names, etc.)
- [ ] Confidence calculation
- [ ] Clarification question generation

#### 5.5 Planner Agent
- [ ] Implement goal decomposition
- [ ] Task breakdown generation
- [ ] Phase suggestion
- [ ] Milestone identification
- [ ] Dependency suggestion

#### 5.6 Executor Agent
- [ ] Permission validation before execution
- [ ] Draft entity creation
- [ ] Action staging
- [ ] Batch execution support

#### 5.7 Human-in-the-Loop
- [ ] Create draft review API
- [ ] Implement approval workflow
- [ ] Add edit capabilities for drafts
- [ ] Implement rejection with feedback

---

### Phase 6: Frontend MVP 🔄 IN PROGRESS (30%)

#### 6.1 Next.js Setup ✅ COMPLETE
- [x] Initialize Next.js 16 with App Router (`npx create-next-app@latest`)
- [x] Turbopack enabled by default (stable in v16)
- [x] Configure TypeScript with strict mode
- [x] Setup Tailwind CSS v4 with CSS-first configuration
- [x] Install and configure shadcn/ui components
- [ ] Setup TanStack Query v5 with React 19 Suspense
- [ ] Configure Zustand stores with devtools
- [ ] Setup React Hook Form with Zod validation
- [ ] Configure Partial Prerendering (PPR) for dynamic routes
- [ ] Leverage React 19 features (use, Actions, useOptimistic)

#### 6.2 Authentication UI
- [x] Login page
- [x] Signup page
- [ ] Forgot password page
- [ ] OAuth buttons
- [ ] Protected route wrapper

#### 6.3 Layout & Navigation
- [x] App shell with sidebar
- [ ] Workspace switcher
- [x] Navigation menu
- [ ] User menu with settings
- [ ] Breadcrumbs

#### 6.4 Workspace Management
- [ ] Workspace list page
- [ ] Create workspace modal
- [ ] Workspace settings page
- [ ] Member management
- [ ] Invitation flow

#### 6.5 Plan Management (Dual-Mode)
- [ ] Plan list page with filters
- [ ] Create plan modal with options:
  - [ ] Manual form creation
  - [ ] Template selection
  - [ ] AI-assisted creation (redirect to chat)
- [ ] Plan settings page
- [ ] Custom status editor
- [ ] Custom fields editor
- [ ] Plan archive/delete
- [ ] "Ask AI for help" contextual button

#### 6.6 Kanban View
- [ ] Board component with columns
- [ ] Task card component
- [ ] Drag-and-drop with dnd-kit
- [ ] Quick add task
- [ ] Swimlanes (by assignee/phase)

#### 6.7 List View
- [ ] Sortable table
- [ ] Inline editing
- [ ] Bulk actions
- [ ] Column customization
- [ ] Grouping

#### 6.8 Calendar View
- [ ] FullCalendar integration
- [ ] Task display by due date
- [ ] Drag to reschedule
- [ ] Month/week/day views

#### 6.9 Gantt View
- [ ] Gantt chart component
- [ ] Task bars with dependencies
- [ ] Milestone markers
- [ ] Zoom controls
- [ ] Date range navigation

#### 6.10 Task Management (Dual-Mode)
- [ ] Task detail modal/drawer
- [ ] Title, description editing (manual)
- [ ] Status, priority, assignee dropdowns (manual)
- [ ] Due date picker (manual)
- [ ] Subtask list with manual add
- [ ] Dependencies editor (manual)
- [ ] Comments section
- [ ] Activity log
- [ ] Quick actions toolbar:
  - [ ] Manual: Click-to-edit fields
  - [ ] AI: Natural language command bar
- [ ] Bulk operations (manual select + action)

#### 6.11 AI Chat Interface
- [ ] Chat panel component
- [ ] Message input with suggestions
- [ ] AI response display
- [ ] Draft review interface
- [ ] Approval/rejection buttons
- [ ] Loading states

#### 6.12 Collaboration Features
- [ ] Real-time presence indicators
- [ ] Live updates via WebSocket
- [ ] Comment notifications
- [ ] @mentions autocomplete
- [ ] Activity feed

#### 6.13 Settings & Profile
- [ ] User profile page
- [ ] Notification preferences
- [ ] Theme settings (dark mode)
- [ ] API key management

---

### Phase 7: Production Readiness ⏳ PENDING

#### 7.1 Security Hardening
- [ ] Security audit (OWASP top 10)
- [ ] Penetration testing
- [ ] Input sanitization review
- [ ] SQL injection prevention audit
- [ ] XSS prevention audit
- [ ] CSRF protection
- [ ] Content Security Policy
- [ ] Rate limiting tuning
- [ ] Secrets rotation setup

#### 7.2 Performance Optimization
- [ ] Database query optimization
- [ ] Add database indexes
- [ ] Implement connection pooling
- [ ] Add caching layer
- [ ] Frontend bundle optimization
- [ ] Image optimization
- [ ] CDN setup
- [ ] Load testing (k6)

#### 7.3 Monitoring & Alerting
- [ ] Setup Prometheus metrics
- [ ] Create Grafana dashboards
- [ ] Configure alerting rules
- [ ] Setup error tracking (Sentry)
- [ ] Create runbooks

#### 7.4 Documentation
- [ ] API documentation
- [ ] User documentation
- [ ] Admin documentation
- [ ] Architecture documentation
- [ ] Deployment documentation

#### 7.5 Deployment
- [ ] Kubernetes manifests
- [ ] Helm chart (optional)
- [ ] Terraform for infrastructure
- [ ] Database backup automation
- [ ] Disaster recovery plan
- [ ] Blue-green deployment setup

---

## 7. User Journeys

> **Design Principle**: Every action supports BOTH manual and AI-assisted workflows. Users can always choose their preferred method. AI assists but never forces.

### Dual-Mode Interaction Model

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    INTERACTION MODES                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   ┌─────────────────────────────┐   ┌─────────────────────────────┐    │
│   │      MANUAL MODE            │   │      AI-ASSISTED MODE       │    │
│   │                             │   │                             │    │
│   │  • Click "Create Plan"      │   │  • Type "Plan a product     │    │
│   │  • Fill form fields         │   │    launch for Q2"           │    │
│   │  • Add tasks one by one     │   │  • AI generates draft       │    │
│   │  • Set dates manually       │   │  • User reviews/approves    │    │
│   │  • Full control             │   │  • AI suggests, user owns   │    │
│   │                             │   │                             │    │
│   └──────────────┬──────────────┘   └──────────────┬──────────────┘    │
│                  │                                  │                   │
│                  └──────────────┬───────────────────┘                   │
│                                 │                                       │
│                                 ▼                                       │
│                  ┌──────────────────────────────┐                       │
│                  │     SAME DATA MODEL          │                       │
│                  │     SAME PERMISSIONS         │                       │
│                  │     SAME AUDIT TRAIL         │                       │
│                  └──────────────────────────────┘                       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Feature Support Matrix

| Feature | Manual | AI-Assisted |
|---------|--------|-------------|
| Create Plan | Form UI | Natural language |
| Add Tasks | Form/inline edit | NL command or bulk generate |
| Update Task | Click & edit | "Mark design review as done" |
| Assign Task | Dropdown select | "Assign to Sarah" |
| Set Due Date | Date picker | "Due next Friday" |
| View Progress | Dashboard | "How's the project going?" |
| Get Suggestions | — | "What tasks am I missing?" |
| Weekly Summary | — | Auto-generated |

---

### Journey 1A: First-Time Planner (MANUAL Plan Creation)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ JOURNEY 1A: MANUAL PLAN CREATION                                        │
└─────────────────────────────────────────────────────────────────────────┘

    ┌─────────┐
    │  START  │
    └────┬────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│   Sign Up       │────▶│  Verify Email   │
│   (OAuth/Email) │     │                 │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Create/Join     │────▶│  Workspace      │
│ Workspace       │     │  Dashboard      │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
┌─────────────────────────────────────────┐
│         CLICK "CREATE PLAN"              │
│  ┌───────────────────────────────────┐  │
│  │  ○ Start from scratch             │  │
│  │  ○ Use template                   │  │
│  │    • Software Project             │  │
│  │    • Marketing Campaign           │  │
│  │    • Event Planning               │  │
│  │    • Operations                   │  │
│  │  ○ Let AI help me plan            │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         │ User selects "Start from scratch"
         ▼
┌─────────────────────────────────────────┐
│         PLAN CREATION FORM               │
│  ┌───────────────────────────────────┐  │
│  │ Plan Name: [Product Launch Q2   ] │  │
│  │ Type:      [Software        ▼]    │  │
│  │ Start:     [Apr 1, 2024      ]    │  │
│  │ End:       [Jun 30, 2024     ]    │  │
│  │ Description:                      │  │
│  │ [Launch our new mobile app...   ] │  │
│  │                                   │  │
│  │ [Cancel]            [Create Plan] │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         │ Plan created (Active status)
         ▼
┌─────────────────────────────────────────┐
│         EMPTY PLAN VIEW                  │
│  ┌───────────────────────────────────┐  │
│  │ Product Launch Q2                 │  │
│  │                                   │  │
│  │  No tasks yet                     │  │
│  │                                   │  │
│  │  [+ Add Task]  [+ Add Phase]      │  │
│  │                                   │  │
│  │  💡 Tip: Need help? Ask AI to    │  │
│  │     suggest tasks for your plan   │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         │ Click "Add Task"
         ▼
┌─────────────────────────────────────────┐
│         TASK CREATION (INLINE)           │
│  ┌───────────────────────────────────┐  │
│  │ Title: [Design app mockups      ] │  │
│  │ Assignee: [Sarah        ▼]        │  │
│  │ Due: [Apr 15]  Priority: [High ▼] │  │
│  │                                   │  │
│  │ [Cancel]              [Add Task]  │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         │ Repeat for each task
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Invite Team     │────▶│ Start Working   │
│ Members         │     │ (Kanban View)   │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
    ┌─────────┐
    │   END   │
    └─────────┘

Technical Implementation:
├── POST /api/workspaces (create workspace)
├── POST /api/workspaces/:wid/plans (create plan - form data)
├── POST /api/plans/:pid/tasks (create task - form data)
├── No AI/MCP involvement
├── Direct CRUD operations
└── Full audit trail maintained
```

---

### Journey 1B: First-Time Planner (AI-Assisted Plan Creation)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ JOURNEY 1: FIRST-TIME PLANNER                                           │
└─────────────────────────────────────────────────────────────────────────┘

    ┌─────────┐
    │  START  │
    └────┬────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│   Sign Up       │────▶│  Verify Email   │
│   (OAuth/Email) │     │                 │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Create/Join     │────▶│  Workspace      │
│ Workspace       │     │  Dashboard      │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
┌─────────────────┐
│  AI Chat:       │
│  "I need to     │
│  plan a         │
│  product        │
│  launch"        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│  Intent Agent   │────▶│  Planner Agent  │
│  Parses Input   │     │  Generates      │
│  confidence:0.94│     │  Draft Plan     │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
┌─────────────────────────────────────────┐
│         DRAFT REVIEW                     │
│  ┌───────────────────────────────────┐  │
│  │ Product Launch Plan               │  │
│  │                                   │  │
│  │ Phases:                           │  │
│  │ 1. Research & Planning            │  │
│  │ 2. Development                    │  │
│  │ 3. Marketing                      │  │
│  │ 4. Launch                         │  │
│  │                                   │  │
│  │ Tasks: 24 suggested               │  │
│  │ Milestones: 4                     │  │
│  └───────────────────────────────────┘  │
│                                         │
│  [✓ Approve]  [✎ Edit]  [✗ Reject]     │
└─────────────────────────────────────────┘
         │
         │ User clicks "Approve"
         ▼
┌─────────────────┐     ┌─────────────────┐
│  Plan Created   │────▶│  Invite Team    │
│  (Active)       │     │  Members        │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
┌─────────────────┐     ┌─────────────────┐
│  Kanban View    │────▶│  Track          │
│  Tasks Visible  │     │  Progress       │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
    ┌─────────┐
    │   END   │
    └─────────┘

Technical Implementation:
├── OAuth2 flow with JWT tokens
├── POST /api/workspaces
├── POST /api/mcp/intent (AI chat)
├── Intent classification + entity extraction
├── Draft stored with status: "draft"
├── POST /api/plans/:id/approve
├── WebSocket connection for real-time updates
└── POST /api/workspaces/:id/invite
```

### Journey 2: Contributor Task Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ JOURNEY 2: CONTRIBUTOR WORKFLOW                                          │
└─────────────────────────────────────────────────────────────────────────┘

    ┌─────────┐
    │  START  │
    └────┬────┘
         │
         ▼
┌─────────────────┐
│ Receive Email   │
│ Invitation      │
│ "You've been    │
│ invited to..."  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Click Invite    │────▶│ Create Account  │
│ Link            │     │ (or Login)      │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
┌─────────────────┐
│ View Dashboard  │
│                 │
│ "My Tasks" (3)  │
│ • Review design │
│ • Write tests   │
│ • Code review   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Click Task:     │
│ "Review design" │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│         TASK DETAIL                      │
│  ┌───────────────────────────────────┐  │
│  │ Review design mockups             │  │
│  │ Status: To Do    Priority: High   │  │
│  │ Due: Jan 30      Assignee: You    │  │
│  │                                   │  │
│  │ Description:                      │  │
│  │ Review the Figma mockups and      │  │
│  │ provide feedback on the new...    │  │
│  │                                   │  │
│  │ Subtasks:                         │  │
│  │ ☐ Review homepage                 │  │
│  │ ☐ Review dashboard                │  │
│  │ ☐ Add comments in Figma           │  │
│  └───────────────────────────────────┘  │
│                                         │
│  Status: [To Do ▼] → [In Progress]     │
└─────────────────────────────────────────┘
         │
         │ Updates status
         ▼
┌─────────────────┐
│ Work on Task    │
│ ...             │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Hit Blocker     │────▶│ Add Comment     │
│                 │     │ "@PM blocked by │
│                 │     │ missing assets" │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         │ Notification sent
         ▼
┌─────────────────┐
│ Blocker         │
│ Resolved        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Complete Work   │────▶│ Mark Done       │
│                 │     │ Status: Done    │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
    ┌─────────┐
    │   END   │
    └─────────┘

Technical Implementation:
├── Email with magic link token
├── GET /api/users/me/tasks
├── GET /api/tasks/:id
├── PATCH /api/tasks/:id { status: "in_progress" }
├── POST /api/tasks/:id/comments
├── WebSocket broadcasts status change
├── Notification sent via email + in-app
└── Activity logged in audit trail
```

### Journey 3: AI-Assisted Task Creation (Natural Language)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ JOURNEY 3: NATURAL LANGUAGE TASK CREATION                               │
└─────────────────────────────────────────────────────────────────────────┘

    ┌─────────┐
    │  START  │
    └────┬────┘
         │
         ▼
┌─────────────────────────────────────────┐
│         AI CHAT INPUT                    │
│  ┌───────────────────────────────────┐  │
│  │ "Add task to review design by     │  │
│  │  Friday, assign to Sarah"         │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│         INTENT AGENT                     │
│                                         │
│  Input: "Add task to review design by   │
│          Friday, assign to Sarah"       │
│                                         │
│  Extracted Entities:                    │
│  ┌───────────────────────────────────┐  │
│  │ intent_type: ADD_TASK             │  │
│  │ confidence: 0.92                  │  │
│  │ entities:                         │  │
│  │   title: "Review design"          │  │
│  │   due_date: "2024-02-02"          │  │
│  │   assignee: "sarah@company.com"   │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         │ confidence > 0.8
         ▼
┌─────────────────────────────────────────┐
│         DRAFT CONFIRMATION               │
│  ┌───────────────────────────────────┐  │
│  │ I'll create this task:            │  │
│  │                                   │  │
│  │ Title: Review design              │  │
│  │ Due: Friday, Feb 2                │  │
│  │ Assignee: Sarah                   │  │
│  │                                   │  │
│  │ Does this look right?             │  │
│  └───────────────────────────────────┘  │
│                                         │
│  [✓ Create]  [✎ Edit]  [✗ Cancel]      │
└─────────────────────────────────────────┘
         │
         │ User clicks "Create"
         ▼
┌─────────────────┐
│ Task Created    │
│ Notification    │
│ sent to Sarah   │
└────────┬────────┘
         │
         ▼
    ┌─────────┐
    │   END   │
    └─────────┘

Technical Implementation:
├── POST /api/mcp/intent
│   └── Body: { "input": "Add task to..." }
├── Claude API call with prompt template
├── Entity extraction with confidence scoring
├── If confidence < 0.6: ask clarification
├── Draft stored with approval_required: true
├── POST /api/tasks (on approval)
├── Send notification to assignee
└── Audit log: AI_TASK_CREATED
```

### Journey 4: Observer Dashboard View

```
┌─────────────────────────────────────────────────────────────────────────┐
│ JOURNEY 4: OBSERVER DASHBOARD                                           │
└─────────────────────────────────────────────────────────────────────────┘

    ┌─────────┐
    │  START  │
    └────┬────┘
         │
         ▼
┌─────────────────┐
│ Login           │
│ (Observer Role) │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        DASHBOARD VIEW                                    │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                     Product Launch Plan                            │  │
│  │                                                                    │  │
│  │  Progress: ████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░  45%          │  │
│  │                                                                    │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │  │
│  │  │ Total Tasks  │  │  Completed   │  │  Overdue     │             │  │
│  │  │     24       │  │     11       │  │      2       │             │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘             │  │
│  │                                                                    │  │
│  │  Milestones                                                        │  │
│  │  ┌─────────────────────────────────────────────────────────────┐  │  │
│  │  │ ✓ Research Complete    │ → Design Complete    │ Launch      │  │  │
│  │  │     Jan 15             │     Feb 1            │   Mar 1     │  │  │
│  │  └─────────────────────────────────────────────────────────────┘  │  │
│  │                                                                    │  │
│  │  [View Gantt Chart]  [View Activity]  [Export Report]             │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  Note: Edit buttons are disabled for Observer role                      │
└─────────────────────────────────────────────────────────────────────────┘
         │
         │ Click "View Gantt Chart"
         ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        GANTT VIEW                                        │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │ Jan          Feb          Mar          Apr                         │  │
│  │ │            │            │            │                           │  │
│  │ ██████████████           Research & Planning                       │  │
│  │              ████████████████         Development                  │  │
│  │                          ██████████████████████ Marketing          │  │
│  │                                      ◆ Launch                      │  │
│  │ │            │            │            │                           │  │
│  │ ▲                        ▲                                         │  │
│  │ Done                     Today                                     │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  Read-only: Cannot drag tasks to reschedule                             │
└─────────────────────────────────────────────────────────────────────────┘
         │
         ▼
    ┌─────────┐
    │   END   │
    └─────────┘

Technical Implementation:
├── GET /api/plans/:id (with computed metrics)
├── GET /api/plans/:id/milestones
├── Permission check: role === "observer"
├── UI disables edit controls
├── Gantt data: GET /api/plans/:id/gantt
└── Export: GET /api/plans/:id/export?format=pdf
```

### Journey 5: Real-Time Collaboration

```
┌─────────────────────────────────────────────────────────────────────────┐
│ JOURNEY 5: REAL-TIME COLLABORATION                                       │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────┐                    ┌─────────────────────┐
│      USER A         │                    │      USER B         │
│    (Plan Owner)     │                    │   (Contributor)     │
└──────────┬──────────┘                    └──────────┬──────────┘
           │                                          │
           │     ┌────────────────────────────┐      │
           │     │      WEBSOCKET SERVER      │      │
           │     │                            │      │
           ├────▶│  Room: plan_abc123         │◀────┤
           │     │  Subscribers: [A, B]       │      │
           │     │                            │      │
           │     └────────────────────────────┘      │
           │                                          │
           ▼                                          │
┌─────────────────────┐                              │
│ User A moves task   │                              │
│ "Design Review"     │                              │
│ from "To Do" to     │                              │
│ "In Progress"       │                              │
└──────────┬──────────┘                              │
           │                                          │
           │ PATCH /api/tasks/xyz                     │
           │ { status: "in_progress" }                │
           ▼                                          │
┌─────────────────────┐                              │
│ Server:             │                              │
│ 1. Update DB        │                              │
│ 2. Emit event       │                              │
└──────────┬──────────┘                              │
           │                                          │
           │ WebSocket broadcast:                     │
           │ {                                        │
           │   type: "task.updated",                  │
           │   task_id: "xyz",                        │
           │   changes: { status: "in_progress" },   │
           │   actor: "user_a"                        │
           │ }                                        │
           │                                          │
           └─────────────────────────────────────────▶│
                                                      │
                                                      ▼
                                        ┌─────────────────────┐
                                        │ User B sees:        │
                                        │                     │
                                        │ Task "Design Review"│
                                        │ moved instantly     │
                                        │                     │
                                        │ [User A avatar]     │
                                        │ "moved task to      │
                                        │  In Progress"       │
                                        └─────────────────────┘

Technical Implementation:
├── WebSocket connection on page load
├── Join room: plan_<plan_id>
├── On API mutation: emit to room
├── Client receives: optimistic update UI
├── Presence: ping/pong for online status
├── Conflict resolution: last-write-wins (MVP)
└── Activity feed: persisted to activity_log table
```

### Journey 6: Hybrid Mode (Manual + AI Assistance)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ JOURNEY 6: HYBRID MODE - SEAMLESS SWITCHING                             │
└─────────────────────────────────────────────────────────────────────────┘

    ┌─────────┐
    │  START  │
    └────┬────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  User creates plan MANUALLY             │
│  (Using form UI)                        │
│                                         │
│  Plan: "Website Redesign"               │
│  5 tasks added manually                 │
└────────────────┬────────────────────────┘
                 │
                 │ User realizes they might
                 │ be missing important tasks
                 ▼
┌─────────────────────────────────────────┐
│         AI CHAT (Side Panel)            │
│  ┌───────────────────────────────────┐  │
│  │ User: "What tasks am I missing    │  │
│  │        for a website redesign?"   │  │
│  └───────────────────────────────────┘  │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│         AI SUGGESTIONS                   │
│  ┌───────────────────────────────────┐  │
│  │ Based on your plan, consider:     │  │
│  │                                   │  │
│  │ ☐ Content audit                   │  │
│  │ ☐ SEO migration plan              │  │
│  │ ☐ 301 redirect mapping            │  │
│  │ ☐ Browser testing                 │  │
│  │ ☐ Performance benchmarking        │  │
│  │ ☐ Accessibility audit             │  │
│  │                                   │  │
│  │ [Add Selected] [Add All] [Ignore] │  │
│  └───────────────────────────────────┘  │
└────────────────┬────────────────────────┘
                 │
                 │ User selects 3 tasks, clicks "Add Selected"
                 ▼
┌─────────────────────────────────────────┐
│  Tasks added to plan                    │
│  (User can edit details manually)       │
│                                         │
│  Total tasks: 8                         │
└────────────────┬────────────────────────┘
                 │
                 │ Later, user drags task
                 │ to "Done" column MANUALLY
                 ▼
┌─────────────────────────────────────────┐
│         MANUAL STATUS UPDATE             │
│                                         │
│  "Content audit" → Done                 │
│  (Drag & drop in Kanban)                │
└────────────────┬────────────────────────┘
                 │
                 │ User wants quick update
                 ▼
┌─────────────────────────────────────────┐
│         AI CHAT (Quick Command)          │
│  ┌───────────────────────────────────┐  │
│  │ User: "Mark SEO migration and     │  │
│  │        redirect mapping as done"  │  │
│  │                                   │  │
│  │ AI: I'll mark these as complete:  │  │
│  │     • SEO migration plan          │  │
│  │     • 301 redirect mapping        │  │
│  │                                   │  │
│  │     [✓ Confirm] [✗ Cancel]        │  │
│  └───────────────────────────────────┘  │
└────────────────┬────────────────────────┘
                 │
                 │ User confirms
                 ▼
┌─────────────────────────────────────────┐
│  Both methods work seamlessly:          │
│                                         │
│  • Manual: Full control, form-based     │
│  • AI: Natural language, batch ops      │
│  • Switch anytime, no mode toggle       │
│  • Same data, same permissions          │
└────────────────┬────────────────────────┘
                 │
                 ▼
    ┌─────────┐
    │   END   │
    └─────────┘

Technical Implementation:
├── AI chat always available (side panel)
├── SUGGEST_TASKS intent for recommendations
├── Batch task creation from suggestions
├── Same API endpoints for manual & AI
├── AI uses same permission checks
└── User can ignore all AI suggestions
```

---

### Journey 7: Weekly AI Summary

```
┌─────────────────────────────────────────────────────────────────────────┐
│ JOURNEY 6: WEEKLY AI SUMMARY                                             │
└─────────────────────────────────────────────────────────────────────────┘

    ┌─────────┐
    │  START  │
    │ (Cron)  │
    └────┬────┘
         │ Every Monday 9am
         ▼
┌─────────────────┐
│ Worker Job:     │
│ Generate        │
│ Weekly Summary  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│         ANALYST AGENT                    │
│                                         │
│  Context:                               │
│  • Plan: Product Launch                 │
│  • Period: Last 7 days                  │
│  • Tasks completed: 5                   │
│  • Tasks added: 3                       │
│  • Blockers: 2                          │
│  • Overdue: 1                           │
│                                         │
│  Generated Summary:                     │
│  ┌───────────────────────────────────┐  │
│  │ Weekly Progress: Product Launch   │  │
│  │                                   │  │
│  │ ✓ Good progress this week:        │  │
│  │   - 5 tasks completed             │  │
│  │   - Design phase 80% complete     │  │
│  │                                   │  │
│  │ ⚠ Attention needed:              │  │
│  │   - "API Integration" is 3 days   │  │
│  │     overdue                       │  │
│  │   - 2 tasks blocked by            │  │
│  │     external dependencies         │  │
│  │                                   │  │
│  │ 📅 This week:                     │  │
│  │   - 8 tasks due                   │  │
│  │   - Milestone "Design Complete"   │  │
│  │     due Friday                    │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Store Summary   │────▶│ Send Email to   │
│ in Database     │     │ Plan Owners     │
└─────────────────┘     └────────┬────────┘
                                 │
         ┌───────────────────────┘
         ▼
┌─────────────────────────────────────────┐
│         EMAIL / DASHBOARD WIDGET         │
│  ┌───────────────────────────────────┐  │
│  │ 📊 Your Weekly Summary            │  │
│  │                                   │  │
│  │ [View in Dashboard]               │  │
│  │                                   │  │
│  │ Quick Stats:                      │  │
│  │ • 5 tasks completed               │  │
│  │ • 1 overdue task                  │  │
│  │ • Milestone due Friday            │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         ▼
    ┌─────────┐
    │   END   │
    └─────────┘

Technical Implementation:
├── Cron job: 0 9 * * 1 (Mondays 9am)
├── Worker: fetch plans with activity
├── For each plan: build context
├── Claude API: generate summary
├── Store: summaries table
├── Email: SendGrid template
└── Dashboard widget: GET /api/plans/:id/summary
```

---

## 8. Data Model

### Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           DATA MODEL                                     │
└─────────────────────────────────────────────────────────────────────────┘

┌──────────────┐
│    users     │
├──────────────┤
│ id           │─────────────────────────────────┐
│ email        │                                 │
│ name         │                                 │
│ password_hash│                                 │
│ avatar_url   │                                 │
│ created_at   │                                 │
│ updated_at   │                                 │
└──────────────┘                                 │
       │                                         │
       │ owns                                    │
       ▼                                         │
┌──────────────┐       ┌──────────────┐          │
│  workspaces  │       │workspace_    │          │
├──────────────┤       │members       │          │
│ id           │◀──────├──────────────┤          │
│ name         │       │ workspace_id │          │
│ slug         │       │ user_id      │◀─────────┘
│ owner_id     │───────│ role         │
│ settings     │       │ joined_at    │
│ created_at   │       └──────────────┘
│ updated_at   │
└──────────────┘
       │
       │ contains
       ▼
┌──────────────┐
│    plans     │
├──────────────┤
│ id           │◀────────────────────────────────┐
│ workspace_id │                                 │
│ name         │                                 │
│ description  │                                 │
│ domain       │   ┌──────────────┐              │
│ status       │   │   phases     │              │
│ owner_id     │   ├──────────────┤              │
│ start_date   │◀──│ id           │              │
│ end_date     │   │ plan_id      │              │
│custom_statuses   │ name         │              │
│ custom_fields│   │ order        │              │
│ tags         │   │ start_date   │              │
│ created_at   │   │ end_date     │              │
│ updated_at   │   └──────────────┘              │
└──────────────┘                                 │
       │                                         │
       │ contains                                │
       ▼                                         │
┌──────────────┐                                 │
│    tasks     │                                 │
├──────────────┤                                 │
│ id           │◀────────────────┐               │
│ plan_id      │─────────────────┼───────────────┘
│ phase_id     │ (optional)      │
│ parent_id    │─────────────────┘ (subtasks)
│ title        │
│ description  │       ┌──────────────┐
│ status       │       │task_         │
│ priority     │       │dependencies  │
│ assignee_id  │       ├──────────────┤
│ due_date     │◀──────│ task_id      │
│ estimated_hrs│       │ depends_on_id│───────────┐
│ custom_fields│       └──────────────┘           │
│ tags         │                                  │
│ created_at   │◀─────────────────────────────────┘
│ updated_at   │
└──────────────┘
       │
       │ has
       ▼
┌──────────────┐       ┌──────────────┐
│   comments   │       │  milestones  │
├──────────────┤       ├──────────────┤
│ id           │       │ id           │
│ task_id      │       │ plan_id      │
│ user_id      │       │ name         │
│ content      │       │ due_date     │
│ created_at   │       │ status       │
└──────────────┘       │ linked_tasks │
                       └──────────────┘

┌──────────────┐       ┌──────────────┐
│ activity_log │       │  audit_log   │
├──────────────┤       ├──────────────┤
│ id           │       │ id           │
│ workspace_id │       │ timestamp    │
│ plan_id      │       │ user_id      │
│ task_id      │       │ workspace_id │
│ user_id      │       │ intent       │
│ action       │       │ action       │
│ details      │       │ result       │
│ created_at   │       │ status       │
└──────────────┘       └──────────────┘
```

### SQL Schema (Key Tables)

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    avatar_url TEXT,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Workspaces
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Workspace Members
CREATE TABLE workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member', -- owner, admin, member, viewer
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);

-- Plans
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    domain VARCHAR(50) NOT NULL DEFAULT 'generic',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    owner_id UUID NOT NULL REFERENCES users(id),
    start_date DATE,
    end_date DATE,
    custom_statuses JSONB DEFAULT '[
        {"name": "To Do", "color": "#6b7280"},
        {"name": "In Progress", "color": "#3b82f6"},
        {"name": "Done", "color": "#10b981"}
    ]',
    custom_fields JSONB DEFAULT '[]',
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tasks
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    phase_id UUID REFERENCES phases(id) ON DELETE SET NULL,
    parent_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(100) NOT NULL DEFAULT 'To Do',
    priority VARCHAR(50) DEFAULT 'medium',
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date DATE,
    estimated_hours DECIMAL(6,2),
    custom_field_values JSONB DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    position INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Task Dependencies
CREATE TABLE task_dependencies (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, depends_on_id),
    CHECK (task_id != depends_on_id)
);

-- Audit Log (for MCP)
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    user_id UUID REFERENCES users(id),
    workspace_id UUID REFERENCES workspaces(id),
    intent JSONB,
    action JSONB,
    result JSONB,
    status VARCHAR(50) NOT NULL, -- success, failed, blocked
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_plans_workspace ON plans(workspace_id);
CREATE INDEX idx_tasks_plan ON tasks(plan_id);
CREATE INDEX idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX idx_tasks_status ON tasks(plan_id, status);
CREATE INDEX idx_tasks_due_date ON tasks(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX idx_audit_log_user ON audit_log(user_id);
CREATE INDEX idx_audit_log_workspace ON audit_log(workspace_id);

-- Full-text search
CREATE INDEX idx_tasks_search ON tasks USING gin(
    to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, ''))
);
```

---

## 9. API Specification

### Authentication Endpoints

```yaml
/api/auth/signup:
  post:
    summary: Create new user account
    requestBody:
      content:
        application/json:
          schema:
            type: object
            required: [email, password, name]
            properties:
              email: { type: string, format: email }
              password: { type: string, minLength: 8 }
              name: { type: string }
    responses:
      201:
        description: User created
        content:
          application/json:
            schema:
              type: object
              properties:
                user: { $ref: '#/components/schemas/User' }
                token: { type: string }

/api/auth/login:
  post:
    summary: Authenticate user
    requestBody:
      content:
        application/json:
          schema:
            type: object
            required: [email, password]
            properties:
              email: { type: string }
              password: { type: string }
    responses:
      200:
        description: Authentication successful
        content:
          application/json:
            schema:
              type: object
              properties:
                user: { $ref: '#/components/schemas/User' }
                access_token: { type: string }
                refresh_token: { type: string }
```

### Plan Endpoints

```yaml
/api/workspaces/{workspace_id}/plans:
  get:
    summary: List plans in workspace
    parameters:
      - name: status
        in: query
        schema: { type: string, enum: [draft, active, on_hold, completed] }
      - name: domain
        in: query
        schema: { type: string }
      - name: page
        in: query
        schema: { type: integer, default: 1 }
      - name: limit
        in: query
        schema: { type: integer, default: 20, maximum: 100 }
    responses:
      200:
        description: List of plans
        content:
          application/json:
            schema:
              type: object
              properties:
                plans:
                  type: array
                  items: { $ref: '#/components/schemas/Plan' }
                pagination: { $ref: '#/components/schemas/Pagination' }

  post:
    summary: Create new plan
    requestBody:
      content:
        application/json:
          schema:
            type: object
            required: [name]
            properties:
              name: { type: string }
              description: { type: string }
              domain: { type: string, enum: [software, event, ops, collection, generic] }
              start_date: { type: string, format: date }
              end_date: { type: string, format: date }
              template_id: { type: string, format: uuid }
    responses:
      201:
        description: Plan created
        content:
          application/json:
            schema: { $ref: '#/components/schemas/Plan' }
```

### Task Endpoints

```yaml
/api/plans/{plan_id}/tasks:
  get:
    summary: List tasks in plan
    parameters:
      - name: status
        in: query
        schema: { type: string }
      - name: assignee_id
        in: query
        schema: { type: string, format: uuid }
      - name: phase_id
        in: query
        schema: { type: string, format: uuid }
      - name: search
        in: query
        schema: { type: string }
    responses:
      200:
        description: List of tasks

  post:
    summary: Create new task
    requestBody:
      content:
        application/json:
          schema:
            type: object
            required: [title]
            properties:
              title: { type: string }
              description: { type: string }
              status: { type: string }
              priority: { type: string, enum: [low, medium, high, critical] }
              assignee_id: { type: string, format: uuid }
              due_date: { type: string, format: date }
              phase_id: { type: string, format: uuid }
              parent_id: { type: string, format: uuid }

/api/tasks/{task_id}:
  patch:
    summary: Update task
    requestBody:
      content:
        application/json:
          schema:
            type: object
            properties:
              title: { type: string }
              description: { type: string }
              status: { type: string }
              priority: { type: string }
              assignee_id: { type: string, format: uuid, nullable: true }
              due_date: { type: string, format: date, nullable: true }
```

### MCP Endpoints

```yaml
/api/mcp/intent:
  post:
    summary: Process natural language intent
    requestBody:
      content:
        application/json:
          schema:
            type: object
            required: [input]
            properties:
              input: { type: string, description: "Natural language input" }
              context:
                type: object
                properties:
                  plan_id: { type: string, format: uuid }
                  workspace_id: { type: string, format: uuid }
    responses:
      200:
        description: Intent processed
        content:
          application/json:
            schema:
              type: object
              properties:
                intent:
                  $ref: '#/components/schemas/MCPIntentEnvelope'
                draft:
                  type: object
                  description: Draft entity if action proposed
                requires_approval:
                  type: boolean

/api/mcp/intent/{intent_id}/approve:
  post:
    summary: Approve AI-generated action
    responses:
      200:
        description: Action executed
        content:
          application/json:
            schema:
              type: object
              properties:
                result: { type: object }
                audit_id: { type: string, format: uuid }

/api/mcp/tools:
  get:
    summary: List available MCP tools
    responses:
      200:
        description: Tool list
        content:
          application/json:
            schema:
              type: array
              items:
                type: object
                properties:
                  name: { type: string }
                  description: { type: string }
                  parameters: { type: object }
```

---

## 10. AI Integration

### Intent Classification

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    INTENT CLASSIFICATION FLOW                            │
└─────────────────────────────────────────────────────────────────────────┘

User Input: "Plan a product launch for our new mobile app in Q2"
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         SYSTEM PROMPT                                    │
│                                                                         │
│  You are GoPlan AI, a planning assistant. Your task is to convert      │
│  user input into a structured intent schema.                            │
│                                                                         │
│  Rules:                                                                 │
│  - Output must be valid JSON matching the MCPIntentEnvelope schema      │
│  - Never create or update data directly                                 │
│  - If confidence < 0.6, set needs_clarification: true                   │
│  - Extract all relevant entities from the input                         │
│                                                                         │
│  Available intent types:                                                │
│  - CREATE_PLAN: User wants to create a new plan                        │
│  - ADD_TASK: User wants to add a task to existing plan                 │
│  - UPDATE_TASK: User wants to modify an existing task                  │
│  - SUGGEST_TASKS: User wants task suggestions                          │
│  - ASK_SUMMARY: User wants a progress summary                          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         CONTEXT PROMPT                                   │
│                                                                         │
│  User: john@company.com                                                 │
│  Role: Planner                                                          │
│  Workspace: Acme Corp (ID: ws_abc123)                                   │
│  Current Plan: None selected                                            │
│  Recent Activity:                                                       │
│  - Created "Marketing Campaign" plan yesterday                          │
│  - Completed 5 tasks in last week                                       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           OUTPUT                                         │
│                                                                         │
│  {                                                                      │
│    "intent_type": "CREATE_PLAN",                                        │
│    "confidence": 0.94,                                                  │
│    "needs_clarification": false,                                        │
│    "entities": {                                                        │
│      "plan_name": "Mobile App Product Launch",                          │
│      "plan_type": "software",                                           │
│      "goal": "Launch new mobile app",                                   │
│      "timeline": "Q2 2024"                                              │
│    },                                                                   │
│    "proposed_actions": [                                                │
│      {                                                                  │
│        "tool": "plan.create",                                           │
│        "arguments": {                                                   │
│          "name": "Mobile App Product Launch",                           │
│          "domain": "software",                                          │
│          "description": "Product launch plan for new mobile app",       │
│          "start_date": "2024-04-01",                                    │
│          "end_date": "2024-06-30"                                       │
│        }                                                                │
│      }                                                                  │
│    ]                                                                    │
│  }                                                                      │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Prompt Templates

```go
// prompts/create_plan.tmpl
const CreatePlanPrompt = `
You are generating a structured plan for: {{.Goal}}

Plan Type: {{.PlanType}}
Timeline: {{.Timeline}}

Generate a comprehensive plan with:
1. 3-5 phases (if applicable for this plan type)
2. 8-15 tasks distributed across phases
3. 2-4 milestones
4. Suggested dependencies between tasks

Output as JSON:
{
  "plan": {
    "name": "string",
    "description": "string",
    "domain": "{{.PlanType}}",
    "phases": [...],
    "tasks": [...],
    "milestones": [...]
  }
}

Consider industry best practices for {{.PlanType}} projects.
`

// prompts/suggest_tasks.tmpl
const SuggestTasksPrompt = `
Given the following plan context:

Plan: {{.PlanName}}
Domain: {{.Domain}}
Existing Tasks:
{{range .ExistingTasks}}
- {{.Title}} ({{.Status}})
{{end}}

Suggest {{.Limit}} additional tasks that would help complete this plan.
Focus on:
- Missing steps that are commonly needed
- Tasks that should come before/after existing tasks
- Risk mitigation tasks

Output as JSON array of task suggestions.
`
```

### Human-in-the-Loop Workflow

```go
// internal/mcp/agents/executor_agent.go

type DraftEntity struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"` // plan, task, etc.
    Data        map[string]interface{} `json:"data"`
    Intent      MCPIntentEnvelope      `json:"intent"`
    CreatedAt   time.Time              `json:"created_at"`
    ExpiresAt   time.Time              `json:"expires_at"` // 24h TTL
    Status      string                 `json:"status"` // pending, approved, rejected
    ApprovedBy  *string                `json:"approved_by,omitempty"`
    ApprovedAt  *time.Time             `json:"approved_at,omitempty"`
}

func (e *ExecutorAgent) StageDraft(ctx context.Context, intent MCPIntentEnvelope) (*DraftEntity, error) {
    // 1. Validate permissions
    if err := e.permChecker.CanExecute(ctx, intent); err != nil {
        return nil, err
    }

    // 2. Create draft entity
    draft := &DraftEntity{
        ID:        uuid.New().String(),
        Type:      extractEntityType(intent),
        Data:      buildDraftData(intent),
        Intent:    intent,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
        Status:    "pending",
    }

    // 3. Store draft
    if err := e.draftStore.Save(ctx, draft); err != nil {
        return nil, err
    }

    // 4. Audit log
    e.auditLogger.Log(ctx, AuditRecord{
        Action:  "draft_created",
        Intent:  intent,
        DraftID: draft.ID,
    })

    return draft, nil
}

func (e *ExecutorAgent) ApproveDraft(ctx context.Context, draftID string, userID string) (*Result, error) {
    // 1. Load draft
    draft, err := e.draftStore.Get(ctx, draftID)
    if err != nil {
        return nil, err
    }

    // 2. Check not expired
    if time.Now().After(draft.ExpiresAt) {
        return nil, ErrDraftExpired
    }

    // 3. Execute the actual tool
    result, err := e.toolRegistry.ExecuteTool(
        draft.Intent.ProposedActions[0].Tool,
        MCPExecutionContext{UserID: userID, Workspace: draft.Intent.Context.WorkspaceID},
        draft.Intent.ProposedActions[0].Arguments,
    )
    if err != nil {
        return nil, err
    }

    // 4. Update draft status
    draft.Status = "approved"
    draft.ApprovedBy = &userID
    now := time.Now()
    draft.ApprovedAt = &now
    e.draftStore.Update(ctx, draft)

    // 5. Audit log
    e.auditLogger.Log(ctx, AuditRecord{
        Action:  "draft_approved",
        DraftID: draft.ID,
        Result:  result,
        Status:  "success",
    })

    return result, nil
}
```

---

## 11. Security Architecture

### Authentication Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    AUTHENTICATION ARCHITECTURE                           │
└─────────────────────────────────────────────────────────────────────────┘

┌──────────┐      ┌──────────────┐      ┌──────────────┐
│  Client  │      │   API GW     │      │  Auth Svc    │
└────┬─────┘      └──────┬───────┘      └──────┬───────┘
     │                   │                     │
     │ POST /auth/login  │                     │
     │ {email, password} │                     │
     │──────────────────▶│                     │
     │                   │ Validate            │
     │                   │────────────────────▶│
     │                   │                     │ Verify password
     │                   │                     │ (Argon2id)
     │                   │◀────────────────────│
     │                   │ {user, tokens}      │
     │◀──────────────────│                     │
     │ {access_token,    │                     │
     │  refresh_token}   │                     │
     │                   │                     │
     │ GET /api/plans    │                     │
     │ Auth: Bearer xxx  │                     │
     │──────────────────▶│                     │
     │                   │ Verify JWT          │
     │                   │ Extract user_id     │
     │                   │ Check permissions   │
     │                   │────────────────────▶│
     │                   │                     │
     │◀──────────────────│                     │
     │ {plans: [...]}    │                     │

Token Structure (JWT):
{
  "sub": "user_uuid",
  "email": "user@example.com",
  "workspaces": ["ws_1", "ws_2"],
  "iat": 1704067200,
  "exp": 1704070800  // 1 hour
}

Refresh Token:
- Stored in HttpOnly cookie
- 7-day expiry
- Rotated on each use
- Invalidated on logout
```

### Authorization (RBAC)

```go
// Permission definitions
type Permission string

const (
    // Workspace permissions
    PermWorkspaceRead   Permission = "workspace:read"
    PermWorkspaceWrite  Permission = "workspace:write"
    PermWorkspaceAdmin  Permission = "workspace:admin"
    PermWorkspaceDelete Permission = "workspace:delete"

    // Plan permissions
    PermPlanCreate      Permission = "plan:create"
    PermPlanRead        Permission = "plan:read"
    PermPlanWrite       Permission = "plan:write"
    PermPlanDelete      Permission = "plan:delete"

    // Task permissions
    PermTaskCreate      Permission = "task:create"
    PermTaskRead        Permission = "task:read"
    PermTaskWrite       Permission = "task:write"
    PermTaskAssign      Permission = "task:assign"
    PermTaskDelete      Permission = "task:delete"
)

// Role definitions
var RolePermissions = map[string][]Permission{
    "owner": {
        PermWorkspaceRead, PermWorkspaceWrite, PermWorkspaceAdmin, PermWorkspaceDelete,
        PermPlanCreate, PermPlanRead, PermPlanWrite, PermPlanDelete,
        PermTaskCreate, PermTaskRead, PermTaskWrite, PermTaskAssign, PermTaskDelete,
    },
    "admin": {
        PermWorkspaceRead, PermWorkspaceWrite, PermWorkspaceAdmin,
        PermPlanCreate, PermPlanRead, PermPlanWrite, PermPlanDelete,
        PermTaskCreate, PermTaskRead, PermTaskWrite, PermTaskAssign, PermTaskDelete,
    },
    "member": {
        PermWorkspaceRead,
        PermPlanCreate, PermPlanRead, PermPlanWrite,
        PermTaskCreate, PermTaskRead, PermTaskWrite,
    },
    "viewer": {
        PermWorkspaceRead,
        PermPlanRead,
        PermTaskRead,
    },
}
```

### Security Checklist

- [ ] **Input Validation**: Validate all inputs server-side
- [ ] **SQL Injection**: Use parameterized queries (sqlc)
- [ ] **XSS**: Sanitize HTML output, CSP headers
- [ ] **CSRF**: SameSite cookies, CSRF tokens for forms
- [ ] **Rate Limiting**: Per-user and per-IP limits
- [ ] **Password Storage**: Argon2id with salt
- [ ] **Secrets**: Environment variables, never in code
- [ ] **HTTPS**: TLS 1.3 only in production
- [ ] **Headers**: HSTS, X-Content-Type-Options, X-Frame-Options
- [ ] **Audit Logging**: All sensitive operations logged
- [ ] **Session Management**: Secure session handling
- [ ] **Dependency Scanning**: Automated CVE scanning

---

## 12. Observability

### Logging Strategy

```go
// Structured logging with zap
logger.Info("task created",
    zap.String("task_id", task.ID),
    zap.String("plan_id", task.PlanID),
    zap.String("user_id", ctx.UserID),
    zap.String("request_id", ctx.RequestID),
    zap.Duration("duration", time.Since(start)),
)

// Log levels:
// - DEBUG: Detailed debugging info (dev only)
// - INFO: Normal operations (task created, user logged in)
// - WARN: Recoverable issues (rate limit approached)
// - ERROR: Failures requiring attention
// - FATAL: Unrecoverable errors (startup failures)
```

### Metrics (Prometheus)

```yaml
# Key metrics to track
- http_requests_total{method, path, status}
- http_request_duration_seconds{method, path}
- db_query_duration_seconds{query}
- mcp_intent_classification_duration_seconds
- mcp_tool_execution_total{tool, status}
- ai_api_calls_total{provider, model}
- ai_api_latency_seconds{provider}
- websocket_connections_active
- tasks_created_total{workspace}
- plans_created_total{workspace, domain}
```

### Tracing (OpenTelemetry)

```go
// Trace spans for key operations
ctx, span := tracer.Start(ctx, "CreateTask")
defer span.End()

span.SetAttributes(
    attribute.String("plan_id", planID),
    attribute.String("user_id", userID),
)

// Child spans for sub-operations
ctx, dbSpan := tracer.Start(ctx, "db.InsertTask")
// ... database operation
dbSpan.End()
```

### Alerting Rules

```yaml
groups:
  - name: goplan
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected

      - alert: SlowAPIResponses
        expr: histogram_quantile(0.95, http_request_duration_seconds_bucket) > 2
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: API responses are slow

      - alert: AIAPIFailures
        expr: rate(ai_api_calls_total{status="error"}[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: AI API call failures increasing
```

---

## 13. Testing Strategy

### Testing Pyramid

```
                    ┌───────────┐
                    │   E2E     │  5%
                    │  Tests    │
                    ├───────────┤
                    │Integration│  20%
                    │  Tests    │
                    ├───────────┤
                    │   Unit    │  75%
                    │  Tests    │
                    └───────────┘
```

### Unit Tests

```go
// internal/domain/task/task_test.go
func TestTask_SetStatus(t *testing.T) {
    tests := []struct {
        name        string
        current     Status
        next        Status
        expectError bool
    }{
        {"todo to in_progress", StatusTodo, StatusInProgress, false},
        {"in_progress to done", StatusInProgress, StatusDone, false},
        {"done to todo", StatusDone, StatusTodo, true}, // Invalid transition
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            task := NewTask("Test", tt.current)
            err := task.SetStatus(tt.next)

            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.next, task.Status)
            }
        })
    }
}
```

### Integration Tests

```go
// internal/infrastructure/postgres/plan_repo_test.go
func TestPlanRepository_Create(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    repo := NewPlanRepository(db)
    ctx := context.Background()

    // Create plan
    plan := &domain.Plan{
        Name:        "Test Plan",
        WorkspaceID: testWorkspaceID,
        OwnerID:     testUserID,
        Domain:      "software",
    }

    err := repo.Create(ctx, plan)
    require.NoError(t, err)
    require.NotEmpty(t, plan.ID)

    // Verify
    loaded, err := repo.GetByID(ctx, plan.ID)
    require.NoError(t, err)
    assert.Equal(t, plan.Name, loaded.Name)
}
```

### E2E Tests

```typescript
// web/e2e/plan-creation.spec.ts
import { test, expect } from '@playwright/test';

test('create plan via AI chat', async ({ page }) => {
  await page.goto('/login');
  await page.fill('[name=email]', 'test@example.com');
  await page.fill('[name=password]', 'password123');
  await page.click('button[type=submit]');

  await page.waitForURL('/dashboard');

  // Open AI chat
  await page.click('[data-testid=ai-chat-toggle]');
  await page.fill('[data-testid=ai-input]', 'Plan a product launch for Q2');
  await page.press('[data-testid=ai-input]', 'Enter');

  // Wait for AI response
  await expect(page.locator('[data-testid=ai-response]')).toBeVisible();
  await expect(page.locator('[data-testid=draft-preview]')).toContainText('Product Launch');

  // Approve draft
  await page.click('[data-testid=approve-draft]');

  // Verify plan created
  await expect(page.locator('[data-testid=plan-title]')).toContainText('Product Launch');
});
```

---

## 14. Deployment Strategy

### Docker Setup

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /goplan ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /goplan /goplan
COPY --from=builder /app/migrations /migrations

EXPOSE 8080
CMD ["/goplan"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://goplan:goplan@postgres:5432/goplan?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_USER=goplan
      - POSTGRES_PASSWORD=goplan
      - POSTGRES_DB=goplan

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  web:
    build: ./web
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes (Production)

```yaml
# deployments/kubernetes/base/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goplan-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: goplan-api
  template:
    metadata:
      labels:
        app: goplan-api
    spec:
      containers:
        - name: api
          image: goplan/api:latest
          ports:
            - containerPort: 8080
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
          envFrom:
            - secretRef:
                name: goplan-secrets
            - configMapRef:
                name: goplan-config
```

### CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: goplan_test
        ports:
          - 5432:5432
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run tests
        run: go test -race -coverprofile=coverage.txt ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/goplan_test?sslmode=disable
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v3

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          ignore-unfixed: true
          severity: 'CRITICAL,HIGH'

  build:
    needs: [test, lint, security]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker image
        run: docker build -t goplan/api:${{ github.sha }} .
      - name: Push to registry
        if: github.ref == 'refs/heads/main'
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker push goplan/api:${{ github.sha }}
```

---

## 15. Success Metrics

### Product Metrics

| Metric | Target (MVP) | Measurement |
|--------|--------------|-------------|
| Plans created | 100/month | Database count |
| Tasks completed | 500/month | Status = done |
| Active users (WAU) | 50 | Unique logins |
| AI suggestion acceptance rate | 60% | Approved/total |
| Time-to-first-task | < 5 minutes | First task timestamp - signup |
| 30-day retention | 40% | Return visits |

### Technical Metrics

| Metric | Target | SLO |
|--------|--------|-----|
| API availability | 99.9% | Monthly |
| API latency (p95) | < 200ms | Per endpoint |
| Error rate | < 0.1% | Per endpoint |
| AI response time | < 3s | Intent processing |
| WebSocket uptime | 99.9% | Monthly |

### Quality Metrics

| Metric | Target |
|--------|--------|
| Test coverage | > 80% |
| Bug escape rate | < 5 bugs/sprint |
| Security vulnerabilities | 0 critical |
| Technical debt ratio | < 5% |

---

## 16. Enhanced Industry Best Practices

### 16.1 API Design Excellence

#### API Versioning Strategy
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    API VERSIONING STRATEGY                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  URL Path Versioning (Recommended for GoPlan):                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  /api/v1/plans                                                   │   │
│  │  /api/v2/plans  (breaking changes)                              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  Deprecation Policy:                                                    │
│  • Announce deprecation 6 months before removal                        │
│  • Return Deprecation header: Deprecation: true                        │
│  • Return Sunset header: Sunset: Sat, 01 Jan 2026 00:00:00 GMT        │
│  • Maintain 2 versions concurrently (N and N-1)                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Idempotency for Safe Retries
```go
// All mutating endpoints support Idempotency-Key header
// POST /api/v1/plans
// Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000

type IdempotencyRecord struct {
    Key        string    `json:"key"`
    Response   []byte    `json:"response"`
    StatusCode int       `json:"status_code"`
    CreatedAt  time.Time `json:"created_at"`
    ExpiresAt  time.Time `json:"expires_at"` // 24h TTL
}

// Implementation ensures:
// - Same key returns cached response
// - Prevents duplicate plan/task creation
// - Essential for webhook delivery
```

#### GraphQL Gateway (Optional Enhancement)
```graphql
# Alongside REST for flexible frontend queries
type Query {
  plan(id: ID!): Plan
  plans(workspaceId: ID!, filter: PlanFilter): PlanConnection!
  myTasks(status: TaskStatus): [Task!]!
}

type Mutation {
  createPlan(input: CreatePlanInput!): Plan!
  updateTask(id: ID!, input: UpdateTaskInput!): Task!
}

type Subscription {
  taskUpdated(planId: ID!): Task!
  planProgress(planId: ID!): ProgressUpdate!
}
```

#### SDK Generation
```yaml
# Auto-generate client SDKs from OpenAPI spec
sdk-generation:
  languages:
    - typescript  # @goplan/sdk-js
    - python      # goplan-sdk
    - go          # github.com/goplan/sdk-go

  features:
    - Type-safe request/response
    - Auto-retry with exponential backoff
    - Built-in authentication handling
    - WebSocket subscription helpers
```

---

### 16.2 Event-Driven Architecture

#### Event Sourcing for Audit Trail
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    EVENT SOURCING PATTERN                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Instead of storing current state, store all events:                    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ events table                                                     │   │
│  ├─────────────────────────────────────────────────────────────────┤   │
│  │ id │ aggregate_id │ type           │ data          │ timestamp  │   │
│  │ 1  │ plan_123     │ PlanCreated    │ {name:...}    │ 2024-01-01 │   │
│  │ 2  │ plan_123     │ TaskAdded      │ {task_id:...} │ 2024-01-02 │   │
│  │ 3  │ plan_123     │ TaskCompleted  │ {task_id:...} │ 2024-01-05 │   │
│  │ 4  │ plan_123     │ PlanCompleted  │ {}            │ 2024-01-10 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  Benefits:                                                              │
│  • Complete audit history (compliance)                                  │
│  • Time-travel debugging                                               │
│  • Rebuild read models from events                                     │
│  • Analytics on historical data                                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### CQRS (Command Query Responsibility Segregation)
```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CQRS PATTERN                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│              ┌──────────────────┐                                       │
│              │     CLIENT       │                                       │
│              └────────┬─────────┘                                       │
│                       │                                                 │
│         ┌─────────────┴─────────────┐                                  │
│         │                           │                                   │
│         ▼                           ▼                                   │
│  ┌─────────────┐            ┌─────────────┐                            │
│  │  COMMANDS   │            │   QUERIES   │                            │
│  │  (Writes)   │            │   (Reads)   │                            │
│  └──────┬──────┘            └──────┬──────┘                            │
│         │                          │                                    │
│         ▼                          ▼                                    │
│  ┌─────────────┐            ┌─────────────┐                            │
│  │   Write     │   Events   │    Read     │                            │
│  │   Model     │───────────▶│   Model     │                            │
│  │ (Normalized)│            │(Denormalized)│                            │
│  └──────┬──────┘            └──────┬──────┘                            │
│         │                          │                                    │
│         ▼                          ▼                                    │
│  ┌─────────────┐            ┌─────────────┐                            │
│  │ PostgreSQL  │            │   Redis/    │                            │
│  │  (Source)   │            │ Elasticsearch│                            │
│  └─────────────┘            └─────────────┘                            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Webhook System
```yaml
# Webhook configuration per workspace
webhooks:
  endpoints:
    - url: https://customer.com/webhooks/goplan
      events:
        - plan.created
        - task.completed
        - milestone.reached
      secret: whsec_xxx  # HMAC signing

  delivery:
    - Retry: 3 attempts with exponential backoff
    - Timeout: 30 seconds
    - Signature: HMAC-SHA256 in X-GoPlan-Signature header
    - Idempotency: Include X-Webhook-ID for deduplication

  monitoring:
    - Track delivery success rate
    - Alert on repeated failures
    - Provide delivery logs in UI
```

---

### 16.3 Multi-Tenancy & Data Isolation

#### Tenant Isolation Strategy
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    MULTI-TENANCY PATTERN                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Shared Database, Separate Schemas (Recommended for GoPlan):            │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     PostgreSQL                                   │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │   │
│  │  │  tenant_a   │ │  tenant_b   │ │  tenant_c   │               │   │
│  │  │  (schema)   │ │  (schema)   │ │  (schema)   │               │   │
│  │  │             │ │             │ │             │               │   │
│  │  │ • plans     │ │ • plans     │ │ • plans     │               │   │
│  │  │ • tasks     │ │ • tasks     │ │ • tasks     │               │   │
│  │  │ • users     │ │ • users     │ │ • users     │               │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  Row-Level Security (RLS) as additional safeguard:                      │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CREATE POLICY workspace_isolation ON plans                      │   │
│  │  USING (workspace_id = current_setting('app.workspace_id'));    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Database Scaling Strategy
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DATABASE SCALING                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Phase 1: Vertical Scaling + Read Replicas                              │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │           ┌──────────────┐                                       │   │
│  │           │   Primary    │                                       │   │
│  │           │   (Writes)   │                                       │   │
│  │           └──────┬───────┘                                       │   │
│  │                  │ Replication                                   │   │
│  │        ┌─────────┼─────────┐                                    │   │
│  │        ▼         ▼         ▼                                    │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐                        │   │
│  │  │ Replica 1│ │ Replica 2│ │ Replica 3│   (Reads)              │   │
│  │  └──────────┘ └──────────┘ └──────────┘                        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  Phase 2: Horizontal Sharding (if needed)                               │
│  • Shard by workspace_id                                               │
│  • Use Citus or Vitess                                                 │
│  • Keep related data co-located                                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

### 16.4 Resilience Patterns

#### Circuit Breaker
```go
// Protect against cascading failures
import "github.com/sony/gobreaker"

var aiServiceBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "ai-service",
    MaxRequests: 3,                    // Requests in half-open
    Interval:    10 * time.Second,     // Reset interval
    Timeout:     30 * time.Second,     // Time in open state
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        return counts.ConsecutiveFailures > 5
    },
    OnStateChange: func(name string, from, to gobreaker.State) {
        log.Info("circuit breaker state change",
            zap.String("name", name),
            zap.String("from", from.String()),
            zap.String("to", to.String()),
        )
    },
})

// Usage
result, err := aiServiceBreaker.Execute(func() (interface{}, error) {
    return claudeClient.GeneratePlan(ctx, prompt)
})
```

#### Graceful Degradation
```go
// When AI service is down, degrade gracefully
func (s *PlanService) CreatePlan(ctx context.Context, input CreatePlanInput) (*Plan, error) {
    plan, err := s.repo.Create(ctx, input)
    if err != nil {
        return nil, err
    }

    // AI suggestions are best-effort
    suggestions, err := s.aiService.SuggestTasks(ctx, plan)
    if err != nil {
        // Log but don't fail
        log.Warn("AI suggestions unavailable", zap.Error(err))
        // Return plan without suggestions
        return plan, nil
    }

    plan.SuggestedTasks = suggestions
    return plan, nil
}
```

#### Bulkhead Pattern
```go
// Isolate resources to prevent cascade failures
type ResourcePools struct {
    DatabasePool    *pgxpool.Pool  // Max 100 connections
    AIPool          *semaphore.Weighted  // Max 10 concurrent AI calls
    WebhookPool     *semaphore.Weighted  // Max 50 concurrent webhooks
}

func (p *ResourcePools) AcquireAI(ctx context.Context) error {
    return p.AIPool.Acquire(ctx, 1)
}
```

---

### 16.5 Feature Management

#### Feature Flags System
```yaml
# Feature flag configuration
feature_flags:
  storage: redis  # or PostgreSQL

  flags:
    ai_task_suggestions:
      default: false
      rules:
        - condition: workspace.plan == "enterprise"
          value: true
        - condition: user.email ends_with "@goplan.io"
          value: true  # Internal testing

    gantt_view:
      default: true
      percentage: 100  # Fully rolled out

    new_kanban_ui:
      default: false
      percentage: 25   # 25% rollout

    ai_auto_assign:
      default: false
      rules:
        - condition: workspace.id in ["ws_beta1", "ws_beta2"]
          value: true  # Beta customers only
```

#### A/B Testing Framework
```go
// Integration with feature flags for experimentation
type Experiment struct {
    Name       string
    Variants   []Variant
    Metrics    []string  // What to measure
    StartDate  time.Time
    EndDate    time.Time
}

type Variant struct {
    Name       string
    Weight     int  // Percentage
    Config     map[string]interface{}
}

// Example: Test new AI prompt
experiment := &Experiment{
    Name: "ai_prompt_v2",
    Variants: []Variant{
        {Name: "control", Weight: 50, Config: map[string]interface{}{"prompt_version": "v1"}},
        {Name: "treatment", Weight: 50, Config: map[string]interface{}{"prompt_version": "v2"}},
    },
    Metrics: []string{"suggestion_acceptance_rate", "time_to_first_task"},
}
```

---

### 16.6 Compliance & Data Governance

#### GDPR/CCPA Compliance
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DATA PRIVACY COMPLIANCE                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Data Subject Rights Implementation:                                    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ Right to Access (GDPR Art. 15)                                   │   │
│  │ • GET /api/v1/users/me/data-export                              │   │
│  │ • Returns all user data in JSON/CSV                             │   │
│  │ • Includes: profile, plans, tasks, comments, activity           │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ Right to Erasure (GDPR Art. 17)                                  │   │
│  │ • DELETE /api/v1/users/me                                       │   │
│  │ • Anonymizes personal data                                       │   │
│  │ • Retains aggregated analytics                                   │   │
│  │ • Cascades to all owned content                                 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ Data Retention Policy                                            │   │
│  │ • Active data: Indefinite while account active                  │   │
│  │ • Deleted accounts: 30-day recovery, then purge                 │   │
│  │ • Audit logs: 7 years (compliance requirement)                  │   │
│  │ • Analytics: Aggregated, anonymized after 2 years               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### SOC 2 Type II Preparation
```yaml
soc2_controls:
  security:
    - [ ] Encryption at rest (AES-256)
    - [ ] Encryption in transit (TLS 1.3)
    - [ ] Access logging for all systems
    - [ ] Vulnerability scanning (weekly)
    - [ ] Penetration testing (annual)
    - [ ] Security awareness training

  availability:
    - [ ] 99.9% uptime SLA
    - [ ] Disaster recovery plan
    - [ ] Backup verification (monthly)
    - [ ] Incident response runbooks

  confidentiality:
    - [ ] Data classification policy
    - [ ] Access control matrix
    - [ ] Vendor security assessments
    - [ ] NDA with all employees

  processing_integrity:
    - [ ] Input validation
    - [ ] Error handling procedures
    - [ ] Change management process

  privacy:
    - [ ] Privacy policy
    - [ ] Data processing agreements
    - [ ] Cookie consent management
```

---

### 16.7 Accessibility (WCAG 2.1 AA)

#### Accessibility Requirements
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    ACCESSIBILITY CHECKLIST                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Perceivable:                                                           │
│  ☐ All images have alt text                                            │
│  ☐ Color contrast ratio ≥ 4.5:1 (text), ≥ 3:1 (large text)           │
│  ☐ Don't rely on color alone to convey information                     │
│  ☐ Captions for video content                                          │
│                                                                         │
│  Operable:                                                              │
│  ☐ Full keyboard navigation                                            │
│  ☐ No keyboard traps                                                   │
│  ☐ Skip navigation links                                               │
│  ☐ Focus indicators visible                                            │
│  ☐ Drag-and-drop has keyboard alternative                              │
│                                                                         │
│  Understandable:                                                        │
│  ☐ Clear error messages                                                │
│  ☐ Consistent navigation                                               │
│  ☐ Input labels and instructions                                       │
│  ☐ Language attribute set                                              │
│                                                                         │
│  Robust:                                                                │
│  ☐ Valid HTML                                                          │
│  ☐ ARIA landmarks and roles                                            │
│  ☐ Screen reader testing (NVDA, VoiceOver)                             │
│  ☐ Works with browser zoom up to 200%                                  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Keyboard Shortcuts
```typescript
// Keyboard navigation for power users
const keyboardShortcuts = {
  global: {
    'mod+k': 'Open command palette',
    'mod+/': 'Open AI chat',
    'mod+n': 'Create new task',
    'mod+shift+n': 'Create new plan',
    'g then h': 'Go to home',
    'g then p': 'Go to plans',
    '?': 'Show keyboard shortcuts',
  },
  kanban: {
    'j': 'Move focus down',
    'k': 'Move focus up',
    'h': 'Move focus left',
    'l': 'Move focus right',
    'enter': 'Open task detail',
    'e': 'Quick edit',
    'd': 'Mark done',
  },
};
```

---

### 16.8 Internationalization (i18n)

#### Multi-Language Support
```typescript
// Next.js i18n configuration
// next.config.js
module.exports = {
  i18n: {
    locales: ['en', 'es', 'fr', 'de', 'ja', 'zh', 'ar'],
    defaultLocale: 'en',
    localeDetection: true,
  },
};

// Translation structure
// locales/en/common.json
{
  "plan": {
    "create": "Create Plan",
    "status": {
      "draft": "Draft",
      "active": "Active",
      "completed": "Completed"
    }
  },
  "task": {
    "create": "Add Task",
    "priority": {
      "low": "Low",
      "medium": "Medium",
      "high": "High",
      "critical": "Critical"
    }
  },
  "ai": {
    "thinking": "AI is thinking...",
    "suggestion": "AI Suggestion",
    "approve": "Approve",
    "reject": "Reject"
  }
}

// RTL support for Arabic, Hebrew
// Automatic layout flipping with CSS logical properties
.task-card {
  margin-inline-start: 1rem;  /* margin-left in LTR, margin-right in RTL */
  padding-inline-end: 0.5rem;
}
```

---

### 16.9 Performance Budgets

#### Frontend Performance Targets
```yaml
performance_budgets:
  core_web_vitals:
    LCP: 2.5s   # Largest Contentful Paint
    FID: 100ms  # First Input Delay
    CLS: 0.1    # Cumulative Layout Shift
    INP: 200ms  # Interaction to Next Paint

  bundle_sizes:
    initial_js: 150KB  # gzipped
    initial_css: 30KB  # gzipped
    per_route: 50KB    # lazy loaded

  api_targets:
    p50: 50ms
    p95: 200ms
    p99: 500ms

  lighthouse_scores:
    performance: 90
    accessibility: 100
    best_practices: 100
    seo: 100

# CI/CD enforcement
performance_ci:
  - lighthouse-ci (block PR if scores drop)
  - bundle-analyzer (alert if size increases >10%)
  - api-benchmark (alert if p95 increases >20%)
```

---

### 16.10 Chaos Engineering

#### Resilience Testing
```yaml
# Chaos experiments using Chaos Monkey or Litmus
chaos_experiments:

  pod_failure:
    description: "Kill random API pods"
    frequency: daily
    expected_behavior: "No user-visible errors, requests retry"

  database_latency:
    description: "Inject 500ms latency to database"
    frequency: weekly
    expected_behavior: "Degraded performance, no failures"

  ai_service_outage:
    description: "Block all AI API calls"
    frequency: weekly
    expected_behavior: "Graceful degradation, manual mode works"

  redis_failure:
    description: "Kill Redis instance"
    frequency: weekly
    expected_behavior: "Cache miss, fall back to database"

  network_partition:
    description: "Partition between services"
    frequency: monthly
    expected_behavior: "Circuit breakers activate, recovery on reconnect"
```

---

### 16.11 Developer Experience

#### Local Development Excellence
```yaml
# docker-compose.dev.yml with hot-reload
development_setup:
  one_command_start: make dev
  hot_reload: true

  local_services:
    - PostgreSQL with seed data
    - Redis
    - Mailhog (email testing)
    - MinIO (S3-compatible storage)
    - LocalStack (AWS services mock)

  developer_tools:
    - API documentation at /docs
    - GraphQL playground at /graphql
    - Database UI at /adminer
    - Tracing UI at /jaeger
    - Metrics at /metrics

# Makefile commands
commands:
  make dev          # Start all services
  make test         # Run tests
  make lint         # Run linters
  make migrate      # Run migrations
  make seed         # Seed test data
  make generate     # Generate code (sqlc, mocks)
  make api-docs     # Generate OpenAPI docs
```

#### Monorepo Tooling
```yaml
# Using Turborepo for monorepo management
turbo.json:
  pipeline:
    build:
      dependsOn: ["^build"]
      outputs: ["dist/**", ".next/**"]
    test:
      dependsOn: ["build"]
    lint:
      outputs: []
    dev:
      cache: false
      persistent: true

# Package structure
packages:
  - apps/api          # Go backend
  - apps/web          # Next.js frontend
  - apps/worker       # Background jobs
  - packages/ui       # Shared UI components
  - packages/sdk      # TypeScript SDK
  - packages/config   # Shared configs (ESLint, TypeScript)
```

---

### 16.12 Cost Optimization

#### Cloud Cost Management
```yaml
cost_optimization:
  compute:
    - Use spot/preemptible instances for workers
    - Right-size based on metrics
    - Auto-scaling with conservative thresholds
    - Reserved instances for predictable baseline

  database:
    - Connection pooling (reduce instance size)
    - Read replicas only when needed
    - Archive old data to cold storage
    - Use appropriate storage tier (SSD vs HDD)

  ai_api:
    - Cache common AI responses
    - Use smaller models for simple tasks
    - Batch requests where possible
    - Set spending limits per workspace

  monitoring:
    - Sample traces (10% in production)
    - Aggregate old metrics
    - Log retention policies

  cdn:
    - Aggressive caching for static assets
    - Image optimization pipeline
    - Compress all responses
```

---

### 16.13 SLOs/SLIs/Error Budgets

#### Service Level Objectives
```yaml
slos:
  api_availability:
    target: 99.9%
    measurement: successful_requests / total_requests
    window: 30 days
    error_budget: 43.2 minutes/month

  api_latency:
    target: 95% of requests < 200ms
    measurement: request_duration_seconds
    window: 30 days

  data_durability:
    target: 99.999999999%  # 11 nines
    measurement: objects_lost / total_objects
    window: 1 year

  ai_response_time:
    target: 90% of AI requests < 5s
    measurement: ai_request_duration_seconds
    window: 7 days

error_budget_policy:
  when_budget_exhausted:
    - Freeze non-critical deployments
    - Focus team on reliability
    - Post-mortem for major incidents

  when_budget_healthy:
    - Ship new features
    - Run chaos experiments
    - Technical debt reduction
```

---

### Summary: Industry Best Practices Checklist

| Category | Practice | Priority | Status |
|----------|----------|----------|--------|
| API Design | Versioning strategy | High | ☐ |
| API Design | Idempotency keys | High | ☐ |
| API Design | SDK generation | Medium | ☐ |
| API Design | GraphQL gateway | Low | ☐ |
| Architecture | Event sourcing | Medium | ☐ |
| Architecture | CQRS | Medium | ☐ |
| Architecture | Webhook system | High | ☐ |
| Data | Multi-tenant isolation | High | ☐ |
| Data | Read replicas | Medium | ☐ |
| Data | Sharding strategy | Low | ☐ |
| Resilience | Circuit breakers | High | ☐ |
| Resilience | Graceful degradation | High | ☐ |
| Resilience | Bulkhead pattern | Medium | ☐ |
| Features | Feature flags | High | ☐ |
| Features | A/B testing | Medium | ☐ |
| Compliance | GDPR/CCPA | High | ☐ |
| Compliance | SOC 2 Type II | Medium | ☐ |
| Compliance | Data retention | High | ☐ |
| UX | Accessibility (WCAG) | High | ☐ |
| UX | Keyboard shortcuts | Medium | ☐ |
| UX | Internationalization | Medium | ☐ |
| Performance | Performance budgets | High | ☐ |
| Performance | Core Web Vitals | High | ☐ |
| Reliability | Chaos engineering | Medium | ☐ |
| Reliability | SLOs/Error budgets | High | ☐ |
| DevEx | One-command setup | High | ☐ |
| DevEx | Monorepo tooling | Medium | ☐ |
| Cost | Cost optimization | Medium | ☐ |

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| MCP | Model Context Protocol - standardized AI agent interface |
| Intent | Parsed user action request with entities and confidence |
| Draft | AI-generated entity awaiting user approval |
| Plan | Top-level project container |
| Phase | Optional grouping of tasks within a plan |
| Workspace | Multi-tenant container for plans and users |

---

## Appendix B: References

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [12-Factor App](https://12factor.net/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [OpenAPI 3.1 Specification](https://spec.openapis.org/oas/v3.1.0)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Next.js Documentation](https://nextjs.org/docs)

---

*Document Version: 1.1*
*Last Updated: January 2026*
*Author: GoPlan Team*
