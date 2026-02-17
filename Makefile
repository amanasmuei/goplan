.PHONY: help dev build test lint clean docker-up docker-down migrate sqlc migrate-up migrate-down frontend api swagger install

# Default target
help:
	@echo "GoPlan Development Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make dev          - Start development environment (Docker)"
	@echo "  make build        - Build all services"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linters"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start all Docker services"
	@echo "  make docker-down  - Stop all Docker services"
	@echo "  make migrate      - Run database migrations"
	@echo "  make api          - Run API locally (requires DB)"
	@echo "  make frontend     - Run frontend locally"

# Development environment
dev:
	docker-compose up -d postgres redis embedding-service
	@echo "Development services started. Run 'make api' to start the API server."

# Build all services
build:
	cd backend && go build -o bin/api ./cmd/api
	cd frontend && yarn build

# Run tests
test:
	cd backend && go test -v -race -cover ./...
	cd frontend && yarn test

# Run linters
lint:
	cd backend && golangci-lint run
	cd frontend && yarn lint

# Clean build artifacts
clean:
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-rebuild:
	docker-compose up -d --build

# Database migrations
migrate:
	@echo "Migrations are automatically run via init.sql on container start"
	@echo "To reset: docker-compose down -v && docker-compose up -d postgres"

# Run API locally
api:
	cd backend && go run ./cmd/api

# Run frontend locally
frontend:
	cd frontend && yarn dev

# Install dependencies
install:
	cd backend && go mod download
	cd frontend && yarn install

# Generate Swagger docs
swagger:
	cd backend && swag init -g cmd/api/main.go -o docs

# Format code
fmt:
	cd backend && go fmt ./...
	cd frontend && yarn format

# Create a test JWT token (for development)
token:
	@echo "Use this endpoint to generate a test token:"
	@echo "curl -X POST http://localhost:8080/api/v1/auth/dev-token"

# Generate sqlc code
sqlc:
	cd sqlc && sqlc generate

# Run database migrations up
migrate-up:
	migrate -path ./migrations -database "$$DATABASE_URL" up

# Run database migrations down one step
migrate-down:
	migrate -path ./migrations -database "$$DATABASE_URL" down 1
