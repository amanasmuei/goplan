# GoPlan

> **Plan anything. Track everything.**

GoPlan is a universal, AI-powered project management platform with MCP (Model Context Protocol) architecture. It adapts to how you work, supporting software projects, events, operations, marketing campaigns, and more.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black?style=flat&logo=next.js)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Features

- **Universal Planning** - One platform for all project types
- **AI-Native** - MCP-first architecture with Claude integration
- **Multiple Views** - Kanban, List, Calendar, and more
- **Human-in-the-Loop** - AI safety with approval workflows
- **Real-time Collaboration** - Work together seamlessly
- **Enterprise Ready** - RBAC, audit logs, SSO support

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 15+ (or use Docker)

### Installation

```bash
# Clone the repository
git clone https://github.com/goplan/goplan.git
cd goplan

# Install dependencies
make install

# Copy environment file
cp .env.example .env

# Start services (PostgreSQL, Redis, Mailhog)
make dev

# Run database migrations
make migrate-up

# Start the backend
go run ./cmd/server

# In another terminal, start the frontend
cd web && npm run dev
```

### Using Docker

```bash
# Build and start everything
docker compose up -d

# View logs
docker compose logs -f
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend                              │
│                    (Next.js 16 + React)                      │
└─────────────────────────┬───────────────────────────────────┘
                          │ REST API
┌─────────────────────────▼───────────────────────────────────┐
│                      API Layer                               │
│              (Go HTTP Handlers + Middleware)                 │
├──────────────────┬──────────────────┬───────────────────────┤
│   Auth (JWT)     │   Rate Limiting  │   Validation          │
└──────────────────┴────────┬─────────┴───────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                     Service Layer                            │
├─────────────────┬─────────────────┬─────────────────────────┤
│  Plan Service   │  Task Service   │   AI Service            │
└─────────────────┴────────┬────────┴─────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   Repository Layer                           │
│                   (sqlc + pgx/v5)                            │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                     PostgreSQL                               │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                     MCP Server                               │
│              (17 Tools for Claude)                           │
├─────────────────┬─────────────────┬─────────────────────────┤
│ Workspace Tools │  Plan Tools     │  Task Tools             │
│ Milestone Tools │  Activity Tools │  Safety Layer           │
└─────────────────┴─────────────────┴─────────────────────────┘
```

## Project Structure

```
goplan/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/             # REST API handlers & middleware
│   ├── auth/            # JWT authentication
│   ├── claude/          # Claude API integration
│   ├── config/          # Configuration management
│   ├── domain/          # Domain entities
│   ├── logging/         # Structured logging
│   ├── mcp/             # MCP server & tools
│   ├── metrics/         # Prometheus metrics
│   ├── postgres/        # Database repositories
│   ├── ratelimit/       # Rate limiting
│   └── validation/      # Input validation
├── web/                 # Next.js frontend
├── migrations/          # SQL migrations
├── deploy/              # Kubernetes & Helm
├── scripts/             # Deployment scripts
└── sqlc/                # SQL queries
```

## API Endpoints

### REST API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/workspaces` | List workspaces |
| POST | `/api/v1/workspaces` | Create workspace |
| GET | `/api/v1/plans` | List plans |
| POST | `/api/v1/plans` | Create plan |
| GET | `/api/v1/tasks` | List tasks |
| POST | `/api/v1/tasks` | Create task |
| POST | `/api/v1/ai/chat` | Chat with AI |
| POST | `/api/v1/ai/plans/generate` | AI plan generation |

### MCP Tools

GoPlan exposes 17 MCP tools for Claude:

- `workspace.list`, `workspace.get`, `workspace.create`
- `plan.list`, `plan.get`, `plan.create`, `plan.update`
- `task.list`, `task.get`, `task.create`, `task.update`, `task.search`, `task.move`
- `milestone.list`, `milestone.create`, `milestone.update`
- `activity.get`

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://goplan:password@localhost:5432/goplan?sslmode=disable

# Authentication
JWT_SECRET=your-secret-key-min-32-characters

# AI (Claude)
CLAUDE_API_KEY=sk-ant-xxx
AI_ENABLED=true
```

See [.env.example](.env.example) for all options.

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Generate sqlc code
make sqlc

# Run with hot reload
air
```

## Deployment

### Kubernetes

```bash
# Deploy with Helm
helm install goplan ./deploy/helm/goplan \
  --namespace goplan \
  --create-namespace \
  -f deploy/helm/goplan/values-production.yaml
```

### Manual

```bash
# Build and deploy
./scripts/deploy.sh production v1.0.0
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
