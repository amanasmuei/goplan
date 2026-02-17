#!/bin/sh
set -e

# Construct DATABASE_URL from individual vars if not already set
if [ -z "$DATABASE_URL" ] && [ -n "$DB_HOST" ]; then
  DATABASE_URL="postgres://${DB_USER:-goplan}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT:-5432}/${DB_NAME:-goplan}?sslmode=${DB_SSLMODE:-disable}"
  export DATABASE_URL
fi

# Run database migrations
if [ -n "$DATABASE_URL" ]; then
  echo "Running database migrations..."
  if [ "$ENVIRONMENT" = "production" ]; then
    # In production, fail fast on migration errors
    migrate -path /app/migrations -database "$DATABASE_URL" up
  else
    # In development, warn but continue
    migrate -path /app/migrations -database "$DATABASE_URL" up || {
      echo "Warning: Migration failed, continuing with startup..."
    }
  fi
else
  echo "Warning: DATABASE_URL not set, skipping migrations"
fi

# Start the application
exec /app/main "$@"
