# GoPlan Contracts Overview

> **IMPORTANT**: All teams MUST read and follow these contracts before implementing.

## Contract Documents

| File | Description | Audience |
|------|-------------|----------|
| `01-domain-types.md` | Entity definitions (Go, TypeScript, SQL) | All teams |
| `02-api-endpoints.md` | REST API specification | Backend, Frontend |
| `03-environment-config.md` | Environment variables | All teams, DevOps |
| `04-mcp-contract.md` | AI/MCP interface specification | Backend, AI |

---

## Quick Reference

### Entity Naming

| Context | Convention | Example |
|---------|------------|---------|
| Go structs | PascalCase | `WorkspaceMember` |
| Go JSON tags | camelCase | `workspaceId` |
| TypeScript | PascalCase (types), camelCase (props) | `interface Plan { planId }` |
| PostgreSQL | snake_case | `workspace_members` |
| API paths | kebab-case (plural nouns) | `/api/v1/workspaces` |

### Core Entities

```
Workspace
  └── Plan
        ├── Phase (optional)
        ├── Milestone
        └── Task
              ├── Subtask (Task with parentId)
              ├── Comment
              └── Dependency
```

### API Base URLs

| Environment | URL |
|-------------|-----|
| Local | `http://localhost:8080/api/v1` |
| Staging | `https://staging-api.goplan.io/api/v1` |
| Production | `https://api.goplan.io/api/v1` |

### Authentication

```
Authorization: Bearer <jwt_token>
```

### Common HTTP Status Codes

| Status | Meaning |
|--------|---------|
| 200 | Success |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 422 | Unprocessable Entity |
| 429 | Rate Limited |
| 500 | Server Error |

---

## Team Responsibilities

### Backend Team (Go)

- [ ] Read `01-domain-types.md` - implement domain entities
- [ ] Read `02-api-endpoints.md` - implement REST handlers
- [ ] Read `03-environment-config.md` - setup configuration
- [ ] Read `04-mcp-contract.md` - implement MCP server

**First deliverables:**
1. Project structure with go.mod
2. Domain entities (internal/domain/*)
3. Database migrations
4. Basic CRUD endpoints

### Frontend Team (Next.js)

- [ ] Read `01-domain-types.md` - create TypeScript types
- [ ] Read `02-api-endpoints.md` - implement API client
- [ ] Read `03-environment-config.md` - setup env vars

**First deliverables:**
1. Project structure with Next.js 16
2. TypeScript types matching contracts
3. API client with type safety
4. Auth flow UI

### Database/DevOps Team

- [ ] Read `01-domain-types.md` - create SQL schema
- [ ] Read `03-environment-config.md` - setup Docker Compose

**First deliverables:**
1. SQL migration files
2. Docker Compose for local dev
3. CI/CD pipeline skeleton

---

## Integration Points

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       INTEGRATION POINTS                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   Frontend (Next.js)                                                    │
│        │                                                                │
│        │ REST API (02-api-endpoints.md)                                │
│        │ WebSocket (real-time updates)                                 │
│        ▼                                                                │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │                     Backend (Go)                                 │  │
│   │                                                                  │  │
│   │   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │  │
│   │   │  REST API   │    │ MCP Server  │    │  WebSocket  │        │  │
│   │   │  Handlers   │    │  (AI)       │    │  Gateway    │        │  │
│   │   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘        │  │
│   │          │                  │                  │                │  │
│   │          └──────────────────┼──────────────────┘                │  │
│   │                             │                                   │  │
│   │                    ┌────────┴────────┐                          │  │
│   │                    │ Domain Services │                          │  │
│   │                    └────────┬────────┘                          │  │
│   │                             │                                   │  │
│   └─────────────────────────────┼───────────────────────────────────┘  │
│                                 │                                      │
│        ┌────────────────────────┼────────────────────────┐            │
│        │                        │                        │            │
│        ▼                        ▼                        ▼            │
│   ┌─────────────┐         ┌─────────────┐         ┌─────────────┐    │
│   │ PostgreSQL  │         │    Redis    │         │  Claude API │    │
│   │ (01-domain) │         │   (cache)   │         │ (04-mcp)    │    │
│   └─────────────┘         └─────────────┘         └─────────────┘    │
│                                                                       │
└───────────────────────────────────────────────────────────────────────┘
```

---

## Sync Points

Teams should sync at these milestones:

| Milestone | What to Verify |
|-----------|----------------|
| **Day 1** | All teams read contracts, clarify questions |
| **Day 3** | Types match across Go, TypeScript, SQL |
| **Day 5** | API client can call backend endpoints |
| **Day 7** | Basic CRUD working end-to-end |

---

## Contract Change Process

1. **Propose** - Create PR with contract changes
2. **Review** - All affected teams review
3. **Approve** - Need approval from each team lead
4. **Update** - Update implementations in all codebases
5. **Test** - Verify integration still works

**DO NOT** change contracts without team consensus!

---

## Questions?

If anything is unclear:
1. Check the detailed contract documents
2. Ask in team channel before assuming
3. Document decisions in ADR (Architecture Decision Records)
