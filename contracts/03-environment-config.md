# Environment Configuration Contract

> This document defines all environment variables. All teams MUST use these exact names.

## Environment Files

```
.env.example     # Template (committed to git)
.env.local       # Local development (gitignored)
.env.test        # Test environment (gitignored)
.env.production  # Production (never in git, use secrets manager)
```

---

## Backend (Go) Environment Variables

### Server Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP server port |
| `HOST` | No | `0.0.0.0` | Server host binding |
| `ENV` | Yes | `development` | Environment: `development`, `staging`, `production` |
| `LOG_LEVEL` | No | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | No | `json` | Log format: `json`, `text` |

### Database

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `DATABASE_MAX_CONNS` | No | `25` | Maximum database connections |
| `DATABASE_MIN_CONNS` | No | `5` | Minimum database connections |
| `DATABASE_MAX_IDLE_TIME` | No | `5m` | Max idle connection time |

Example:
```
DATABASE_URL=postgres://goplan:password@localhost:5432/goplan?sslmode=disable
```

### Redis

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REDIS_URL` | Yes | — | Redis connection string |
| `REDIS_PASSWORD` | No | — | Redis password (if required) |
| `REDIS_DB` | No | `0` | Redis database number |

Example:
```
REDIS_URL=redis://localhost:6379
```

### Authentication

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes | — | JWT signing secret (min 32 chars) |
| `JWT_ACCESS_EXPIRY` | No | `1h` | Access token expiry |
| `JWT_REFRESH_EXPIRY` | No | `7d` | Refresh token expiry |
| `BCRYPT_COST` | No | `12` | Bcrypt hashing cost |

### OAuth Providers (Optional)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OAUTH_GOOGLE_CLIENT_ID` | No | — | Google OAuth client ID |
| `OAUTH_GOOGLE_CLIENT_SECRET` | No | — | Google OAuth client secret |
| `OAUTH_GITHUB_CLIENT_ID` | No | — | GitHub OAuth client ID |
| `OAUTH_GITHUB_CLIENT_SECRET` | No | — | GitHub OAuth client secret |

### AI / LLM

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CLAUDE_API_KEY` | Yes* | — | Anthropic Claude API key |
| `CLAUDE_MODEL` | No | `claude-sonnet-4-20250514` | Claude model to use |
| `AI_MAX_TOKENS` | No | `4096` | Max tokens per request |
| `AI_TIMEOUT` | No | `30s` | AI request timeout |
| `AI_ENABLED` | No | `true` | Enable/disable AI features |

*Required only if AI features are enabled

### Email

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SMTP_HOST` | No | — | SMTP server host |
| `SMTP_PORT` | No | `587` | SMTP server port |
| `SMTP_USER` | No | — | SMTP username |
| `SMTP_PASSWORD` | No | — | SMTP password |
| `SMTP_FROM` | No | `noreply@goplan.io` | From email address |
| `SENDGRID_API_KEY` | No | — | SendGrid API key (alternative to SMTP) |

### Storage

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STORAGE_TYPE` | No | `local` | Storage type: `local`, `s3` |
| `STORAGE_PATH` | No | `./uploads` | Local storage path |
| `S3_BUCKET` | No | — | S3 bucket name |
| `S3_REGION` | No | — | S3 region |
| `S3_ACCESS_KEY` | No | — | S3 access key |
| `S3_SECRET_KEY` | No | — | S3 secret key |
| `S3_ENDPOINT` | No | — | S3 endpoint (for MinIO) |

### Security

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CORS_ORIGINS` | No | `*` | Allowed CORS origins (comma-separated) |
| `RATE_LIMIT_ENABLED` | No | `true` | Enable rate limiting |
| `RATE_LIMIT_REQUESTS` | No | `100` | Requests per window |
| `RATE_LIMIT_WINDOW` | No | `1m` | Rate limit window |

### Observability

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OTEL_ENABLED` | No | `false` | Enable OpenTelemetry |
| `OTEL_ENDPOINT` | No | — | OTLP endpoint |
| `OTEL_SERVICE_NAME` | No | `goplan-api` | Service name for tracing |
| `SENTRY_DSN` | No | — | Sentry DSN for error tracking |

---

## Frontend (Next.js) Environment Variables

### Public Variables (Exposed to Browser)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | Yes | — | Backend API URL |
| `NEXT_PUBLIC_WS_URL` | Yes | — | WebSocket URL |
| `NEXT_PUBLIC_APP_NAME` | No | `GoPlan` | Application name |
| `NEXT_PUBLIC_APP_URL` | Yes | — | Frontend URL |

Example:
```
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

### Server-Only Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `API_SECRET` | No | — | Internal API secret |

### Analytics (Optional)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXT_PUBLIC_GA_ID` | No | — | Google Analytics ID |
| `NEXT_PUBLIC_POSTHOG_KEY` | No | — | PostHog API key |
| `NEXT_PUBLIC_POSTHOG_HOST` | No | — | PostHog host |

---

## Docker Compose Variables

For local development, create `.env` in project root:

```bash
# .env (for docker-compose)

# PostgreSQL
POSTGRES_USER=goplan
POSTGRES_PASSWORD=goplan_dev_password
POSTGRES_DB=goplan

# Redis
REDIS_PASSWORD=

# Backend
DATABASE_URL=postgres://goplan:goplan_dev_password@postgres:5432/goplan?sslmode=disable
REDIS_URL=redis://redis:6379
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
CLAUDE_API_KEY=sk-ant-xxx

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

---

## Example .env.example File

```bash
# ===========================================
# GoPlan Environment Configuration
# Copy this file to .env.local and fill in values
# ===========================================

# ----- Server -----
PORT=8080
ENV=development
LOG_LEVEL=debug

# ----- Database -----
DATABASE_URL=postgres://goplan:password@localhost:5432/goplan?sslmode=disable

# ----- Redis -----
REDIS_URL=redis://localhost:6379

# ----- Authentication -----
JWT_SECRET=change-this-to-a-secure-secret-min-32-characters
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=7d

# ----- AI (Claude) -----
CLAUDE_API_KEY=sk-ant-xxx
CLAUDE_MODEL=claude-sonnet-4-20250514
AI_ENABLED=true

# ----- Email (Optional) -----
# SENDGRID_API_KEY=SG.xxx

# ----- OAuth (Optional) -----
# OAUTH_GOOGLE_CLIENT_ID=
# OAUTH_GOOGLE_CLIENT_SECRET=
# OAUTH_GITHUB_CLIENT_ID=
# OAUTH_GITHUB_CLIENT_SECRET=

# ----- Frontend -----
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

---

## Configuration Loading Priority

### Backend (Go with Viper)
1. Environment variables (highest priority)
2. `.env` file
3. Default values

### Frontend (Next.js)
1. `.env.local` (highest priority)
2. `.env.development` / `.env.production`
3. `.env`

---

## Secrets Management (Production)

**DO NOT commit secrets to git!**

For production, use:
- AWS Secrets Manager
- HashiCorp Vault
- Kubernetes Secrets
- Cloud provider secret management

```yaml
# Kubernetes Secret example
apiVersion: v1
kind: Secret
metadata:
  name: goplan-secrets
type: Opaque
stringData:
  DATABASE_URL: "postgres://..."
  JWT_SECRET: "..."
  CLAUDE_API_KEY: "..."
```
