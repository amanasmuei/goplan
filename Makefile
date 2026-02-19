.PHONY: help dev build test lint clean docker-up docker-down docker-logs docker-rebuild api frontend swagger install fmt token migrate-up migrate-down

# Default target
help:
	@echo "GoPlan Development Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make dev          - Start dev services (postgres, redis, embedding-service)"
	@echo "  make api          - Run Fiber API locally (requires DB)"
	@echo "  make frontend     - Run Next.js frontend locally"
	@echo "  make build        - Build backend binary + frontend production build"
	@echo "  make test         - Run backend + frontend tests"
	@echo "  make lint         - Run backend + frontend linters"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start all Docker services"
	@echo "  make docker-down  - Stop all Docker services"
	@echo "  make swagger      - Regenerate Swagger docs"
	@echo "  make install      - Download Go + Node dependencies"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback one migration"

# Development environment
dev:
	docker-compose up -d postgres redis embedding-service
	@echo "Development services started. Run 'make api' to start the API server."

# Run API locally
api:
	cd backend && go run ./cmd/api

# Run frontend locally
frontend:
	cd frontend-next && yarn dev

# Build all services
build:
	cd backend && go build -o bin/api ./cmd/api
	cd frontend-next && yarn build

# Run tests
test:
	cd backend && go test -v -race -cover ./...
	cd frontend-next && yarn test

# Run linters
lint:
	cd backend && golangci-lint run
	cd frontend-next && yarn lint

# Clean build artifacts
clean:
	rm -rf backend/bin
	rm -rf frontend-next/.next
	rm -rf frontend-next/node_modules

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-rebuild:
	docker-compose up -d --build

# Install dependencies
install:
	cd backend && go mod download
	cd frontend-next && yarn install

# Generate Swagger docs
swagger:
	cd backend && swag init -g cmd/api/main.go -o docs

# Format code
fmt:
	cd backend && go fmt ./...

# Create a test JWT token (for development)
token:
	@echo "Use this endpoint to generate a test token:"
	@echo "curl -X POST http://localhost:8080/api/v1/auth/dev-token"

# Run database migrations up
migrate-up:
	migrate -path ./backend/migrations -database "$$DATABASE_URL" up

# Run database migrations down one step
migrate-down:
	migrate -path ./backend/migrations -database "$$DATABASE_URL" down 1
