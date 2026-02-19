# GoPlan

> **Turn any idea into a structured strategic plan.**

GoPlan is an AI-powered strategy consultant platform. Input your idea or goal and receive a consultant-grade strategic plan — not a chat response, but a structured 5-section framework covering executive brief, strategic context, recommended approach, phased execution, and immediate actions.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black?style=flat&logo=next.js)](https://nextjs.org/)
[![Claude AI](https://img.shields.io/badge/Claude-Anthropic-blueviolet?style=flat)](https://anthropic.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## What It Does

You describe what you want to achieve. GoPlan returns a structured strategic plan:

```
"I want to launch a SaaS product for restaurant inventory management"
```

**Result:** A 5-section strategic framework:

| Section | What You Get |
|---------|-------------|
| Executive Brief | Summary, core objective, scope, key stakeholders |
| Strategic Context | Market analysis, opportunities, threats, assumptions |
| Recommended Approach | Core strategy, pillars, alternatives, risk assessment |
| Phased Execution | Timeline with phases, milestones, and concrete actions |
| Immediate Action Plan | Quick wins, critical path, week-by-week plan |

Each section is individually refinable — regenerate with new context or provide feedback to improve specific parts.

## Features

- **AI Strategy Generation** — Single Claude API call produces all 5 sections via a 3-layer prompt (intent classification + consultant framework + category overlay)
- **8 Plan Categories** — Business, SaaS, Event, Marketing, Operations, Product, Research, Personal — each with domain-specific prompt intelligence
- **Section Refinement** — Regenerate or refine individual sections with additional context
- **Version History** — Every regeneration creates a version; browse and compare past iterations
- **Similar Strategies** — pgvector-powered semantic search finds related plans in your history
- **Export** — Download plans as Markdown or JSON
- **Subscription Tiers** — Free (2 plans), Pro ($49/mo), Pro Plus ($99/mo) with feature gating
- **3-Panel Strategy Viewer** — Desktop-first layout: section navigation, content display, and action panel

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 20+ & Yarn
- Docker & Docker Compose
- PostgreSQL 15+ with pgvector (or use Docker)

### Installation

```bash
# Clone the repository
git clone https://github.com/goplan/goplan.git
cd goplan

# Install dependencies
make install

# Copy environment file
cp .env.example .env
# Edit .env — set CLAUDE_API_KEY, JWT_SECRET, database credentials

# Start services (PostgreSQL with pgvector, Redis, embedding service)
make dev

# Start the backend API
make api

# In another terminal, start the frontend
make frontend
```

### Using Docker

```bash
# Build and start everything
docker compose up -d

# View logs
docker compose logs -f
```

The API is available at `http://localhost:8080` and the frontend at `http://localhost:3000`.

Swagger docs: `http://localhost:8080/swagger/index.html` (development only).

## User Flow

```
Register/Login
      │
      ▼
  Dashboard ──────── View existing plans (card grid)
      │
      ▼
  New Strategy ───── Describe your idea (20-5000 chars)
      │                Select category (optional)
      ▼
  AI Generation ──── Claude generates 5 structured sections (~10s)
      │
      ▼
  Strategy Viewer ── 3-panel layout:
      │                ├── Left:   Section navigation
      │                ├── Center: Structured content display
      │                └── Right:  Actions (regenerate, refine, export)
      │
      ▼
  Iterate ────────── Refine sections, regenerate with new context
      │                Version history tracks all changes
      ▼
  Export ──────────── Markdown or JSON download
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (Next.js 16)                    │
│          App Router · SSR · shadcn/ui · TanStack Query       │
└──────────────────────────┬──────────────────────────────────┘
                           │ REST API
┌──────────────────────────▼──────────────────────────────────┐
│                     Fiber API (Go)                           │
├──────────┬───────────┬──────────────┬───────────────────────┤
│ JWT Auth │ Rate Limit│ Subscription │ Security Headers       │
│          │           │ Middleware   │                         │
└──────────┴─────┬─────┴──────────────┴───────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│                    Service Layer                             │
├────────────────┬────────────────┬───────────────────────────┤
│ Strategy       │ Prompt Builder │ Embedding Client           │
│ Service        │ (3-layer AI)   │ (pgvector)                 │
└────────┬───────┴────────┬───────┴───────────┬───────────────┘
         │                │                   │
    ┌────▼────┐    ┌──────▼──────┐    ┌───────▼───────┐
    │ Claude  │    │ Repository  │    │  Embedding    │
    │ API     │    │ Layer       │    │  Service      │
    │(Anthropic)   │ (pgx/v5)   │    │  (FastAPI)    │
    └─────────┘    └──────┬──────┘    └───────────────┘
                          │
                   ┌──────▼──────┐
                   │ PostgreSQL  │
                   │ + pgvector  │
                   └─────────────┘
```

## Project Structure

```
goplan/
├── backend/
│   ├── cmd/api/main.go              # Entry point, route registration
│   ├── internal/
│   │   ├── claude/                   # Anthropic API client (streaming, rate limiting)
│   │   ├── config/                   # Environment configuration
│   │   ├── database/                 # PostgreSQL connection pool
│   │   ├── handlers/
│   │   │   ├── auth_handler.go       # Login, register, profile
│   │   │   ├── strategy_handler.go   # 11 strategy endpoints
│   │   │   └── subscription_handler.go
│   │   ├── middleware/
│   │   │   ├── auth.go               # JWT authentication
│   │   │   └── subscription.go       # Tier-based feature gating
│   │   ├── models/
│   │   │   ├── strategic_plan.go     # Plan entity, filters, enums
│   │   │   ├── plan_section.go       # 5 section types, ordering
│   │   │   ├── section_content.go    # Typed content structs per section
│   │   │   ├── subscription.go       # Tier definitions, limits
│   │   │   └── version.go            # Section & plan versioning
│   │   ├── repository/
│   │   │   ├── plan_repository.go    # CRUD, embedding, similarity search
│   │   │   ├── section_repository.go # Batch create, by type lookup
│   │   │   ├── version_repository.go # Section & plan version history
│   │   │   ├── subscription_repository.go
│   │   │   └── generation_log_repository.go
│   │   ├── services/
│   │   │   ├── strategy_service.go   # Core AI: generate, regenerate, refine
│   │   │   ├── prompt_builder.go     # 3-layer prompt construction
│   │   │   └── embedding_client.go   # HTTP client for embedding service
│   │   └── workers/
│   │       └── strategy_embedding_worker.go  # Background pgvector indexing
│   ├── migrations/                   # SQL migrations (golang-migrate)
│   └── docs/                         # Swagger generated docs
├── frontend-next/
│   ├── src/
│   │   ├── app/
│   │   │   ├── page.tsx              # Landing page (SSR, SEO)
│   │   │   ├── (auth)/               # Login & register
│   │   │   └── (dashboard)/
│   │   │       ├── dashboard/        # Plan grid
│   │   │       └── strategies/
│   │   │           ├── new/          # Strategy input form
│   │   │           └── [id]/         # 3-panel viewer + versions
│   │   ├── components/
│   │   │   ├── strategy/             # SectionNav, SectionViewer, ActionPanel
│   │   │   │   └── SectionContent/   # 5 section-specific renderers
│   │   │   ├── layout/               # Sidebar, Navbar
│   │   │   └── auth/                 # AuthGuard
│   │   ├── hooks/                    # useStrategy, useAuth, useSubscription
│   │   ├── lib/                      # API client, types, constants
│   │   └── store/                    # Zustand auth store
│   └── public/
├── docker-compose.yml
├── Makefile
├── .env.example
└── CLAUDE.md
```

## API Endpoints

### Auth

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Create account |
| POST | `/api/v1/auth/login` | Login, returns JWT |
| GET | `/api/v1/auth/me` | Current user profile |

### Strategies

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/strategies` | Generate a new strategic plan |
| GET | `/api/v1/strategies` | List user's plans (paginated) |
| GET | `/api/v1/strategies/:id` | Get plan with all sections |
| DELETE | `/api/v1/strategies/:id` | Archive a plan |
| POST | `/api/v1/strategies/:id/sections/:type/regenerate` | Regenerate a section (Pro) |
| POST | `/api/v1/strategies/:id/sections/:type/refine` | Refine with feedback (Pro) |
| GET | `/api/v1/strategies/:id/versions` | Plan version history (Pro) |
| GET | `/api/v1/strategies/:id/versions/:version` | Get specific version |
| GET | `/api/v1/strategies/:id/sections/:type/versions` | Section version history |
| GET | `/api/v1/strategies/:id/similar` | Similar strategies (pgvector) |
| GET | `/api/v1/strategies/:id/export` | Export as Markdown or JSON (Pro) |

### Subscription

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/subscription` | Current subscription |
| POST | `/api/v1/subscription/upgrade` | Upgrade tier |

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
# Server
PORT=8080
ENVIRONMENT=development
ALLOW_ORIGINS=http://localhost:3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=goplan
DB_PASSWORD=password
DB_NAME=goplan
DB_SSLMODE=disable

# Authentication
JWT_SECRET=your-secret-key-min-32-characters
JWT_EXPIRY=24h

# AI (Claude) — required for strategy generation
AI_ENABLED=true
CLAUDE_API_KEY=sk-ant-xxx
CLAUDE_MODEL=claude-sonnet-4-20250514
AI_MAX_TOKENS=8192
AI_TIMEOUT_SEC=120
AI_RATE_LIMIT_RPM=50

# Embedding Service (optional — enables similarity search)
EMBEDDING_SERVICE_URL=http://localhost:8000

# Stripe (optional — for paid subscriptions)
STRIPE_SECRET_KEY=sk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
```

See [.env.example](.env.example) for all options.

## Development

```bash
# Run backend tests
cd backend && go test -v -cover ./...

# Run frontend type check
cd frontend-next && yarn type-check

# Lint
make lint

# Regenerate Swagger docs
make swagger

# Run backend with hot reload
cd backend && air
```

## Subscription Tiers

| Feature | Free | Pro ($49/mo) | Pro Plus ($99/mo) |
|---------|------|-------------|-------------------|
| Strategic plans | 2 | 50 | Unlimited |
| Regenerations/day | 0 | 10 | Unlimited |
| Refinement | - | Yes | Yes |
| Version history | - | Yes | Yes |
| Export | - | Yes | Yes |
| Similar strategies | - | - | Yes |
| Priority generation | - | - | Yes |

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24, Fiber v2 |
| Frontend | Next.js 16, React 19, TypeScript |
| UI | Tailwind CSS v4, shadcn/ui |
| State | Zustand (auth), TanStack Query (server) |
| Database | PostgreSQL 15+ with pgvector |
| AI | Claude (Anthropic API) |
| Embeddings | FastAPI Python service |
| Cache | Redis 7 |
| Docs | Swagger/OpenAPI |

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
