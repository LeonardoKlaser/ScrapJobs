#!/bin/sh
set -e

# Run migrations only for the API service
if [ "$SERVICE_CMD" = "/app/api" ] && [ -n "$DATABASE_URL" ]; then
  echo "Running database migrations..."
  /app/migrate -path /app/migrations -database "$DATABASE_URL" up || true
  echo "Migrations complete."
fi

exec $SERVICE_CMD
