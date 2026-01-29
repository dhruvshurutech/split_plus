#!/bin/sh
set -e

echo "Waiting for database to be ready..."
# Wait for postgres to be ready (healthcheck ensures this, but adding a small delay)
sleep 2

echo "Running database migrations..."
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

echo "Migrations completed. Starting server..."
exec air -c .air.toml
