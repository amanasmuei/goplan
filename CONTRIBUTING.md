# Contributing to GoPlan

Thank you for your interest in contributing to GoPlan! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and constructive in all interactions.

## How to Contribute

### Reporting Bugs

1. Check existing issues to avoid duplicates
2. Use the bug report template
3. Include:
   - Clear description of the bug
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)

### Suggesting Features

1. Check existing issues and discussions
2. Use the feature request template
3. Describe the use case and proposed solution

### Pull Requests

1. **Fork** the repository
2. **Clone** your fork locally
3. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. **Make your changes** following our coding standards
5. **Write tests** for new functionality
6. **Run tests** to ensure nothing is broken:
   ```bash
   make test
   make lint
   ```
7. **Commit** with a descriptive message:
   ```bash
   git commit -m "feat: add new feature description"
   ```
8. **Push** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
9. **Open a Pull Request** against the `main` branch

## Development Setup

### Prerequisites

- Go 1.23+
- Node.js 20+
- Docker & Docker Compose
- Make

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/goplan.git
cd goplan

# Install dependencies
make install

# Copy environment file
cp .env.example .env

# Start development services
make dev

# Run migrations
make migrate-up

# Start the backend
go run ./cmd/server

# Start the frontend (in another terminal)
cd web && npm run dev
```

## Coding Standards

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write table-driven tests
- Add comments for exported functions

### TypeScript/React

- Use TypeScript for all new code
- Follow the existing code style
- Use functional components with hooks
- Add proper types (avoid `any`)

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

### Branch Naming

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation
- `refactor/` - Code refactoring

## Testing

### Backend

```bash
# Run all tests
make test

# Run tests with coverage
make test-backend-coverage

# Run specific package tests
go test ./internal/auth/...
```

### Frontend

```bash
cd web
npm run test
npm run test:coverage
```

## Documentation

- Update README.md for user-facing changes
- Add inline comments for complex logic
- Update API documentation for endpoint changes

## Questions?

- Open a [Discussion](https://github.com/goplan/goplan/discussions)
- Check existing issues and discussions first

Thank you for contributing!
