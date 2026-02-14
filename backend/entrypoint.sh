#!/bin/sh
set -e

# Run database migrations if DATABASE_URL is set
if [ -n "$DATABASE_URL" ]; then
  echo "Running database migrations..."
  migrate -path /app/migrations -database "$DATABASE_URL" up || {
    echo "Warning: Migration failed, continuing with startup..."
  }
fi

# Start the application
exec /app/main "$@"
