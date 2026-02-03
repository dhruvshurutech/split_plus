# Split+ Project Commands
# Run `just` to see all available commands

# Default recipe - show help
default:
    @just --list

# ============================================================================
# Docker Commands
# ============================================================================

# Start all services (frontend, backend, worker, postgres)
up:
    @echo "Starting Docker services..."
    set -a; [ -f be/.env ] && source be/.env; set +a; docker compose -f docker/docker-compose.yml up -d

# Stop all services
down *FLAGS:
    @echo "Stopping Docker services..."
    set -a; [ -f be/.env ] && source be/.env; set +a; docker compose -f docker/docker-compose.yml down --remove-orphans {{FLAGS}}

# Stop services and remove volumes (clean slate)
down-v:
    @echo "Stopping Docker services and removing volumes..."
    set -a; [ -f be/.env ] && source be/.env; set +a; docker compose -f docker/docker-compose.yml down -v

# Rebuild and restart services
restart:
    @echo "Rebuilding and restarting services..."
    set -a; [ -f be/.env ] && source be/.env; set +a; docker compose -f docker/docker-compose.yml up -d --build

# Connect to postgres database
db:
    set -a; [ -f be/.env ] && source be/.env; set +a; docker compose -f docker/docker-compose.yml exec postgres psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}"

# View logs from all services
logs:
    docker compose -f docker/docker-compose.yml logs -f

# View logs from postgres service only
logs-db:
    docker compose -f docker/docker-compose.yml logs -f postgres

# View logs from frontend service only
logs-frontend:
    docker compose -f docker/docker-compose.yml logs -f frontend

# View logs from backend service only
logs-backend:
    docker compose -f docker/docker-compose.yml logs -f backend

# View logs from worker service only
logs-worker:
    docker compose -f docker/docker-compose.yml logs -f worker

# ============================================================================
# Docker Helper
# ============================================================================

# Open a shell in the backend container
shell:
    docker compose -f docker/docker-compose.yml exec backend sh

# Execute a command in the backend container
exec cmd:
    docker compose -f docker/docker-compose.yml exec backend sh -c "{{cmd}}"

# ============================================================================
# Test Commands (Local)
# ============================================================================

# Run Go unit tests with optional flags
test *FLAGS:
    @echo "Running unit tests..."
    cd be && go test {{FLAGS}} ./...

# Run end-to-end integration tests with test DB
test-e2e:
    @echo "Starting integration test database..."
    docker compose -f docker/docker-compose.test.yml down -v --remove-orphans
    sleep 2
    docker compose -f docker/docker-compose.test.yml up -d
    echo "Waiting for test DB to be ready..."
    sleep 2
    @echo "Loading .env.test if present..."
    set -a; [ -f be/.env.test ] && source be/.env.test; set +a
    @echo "Running integration test suite..."
    DATABASE_URL_TEST="${DATABASE_URL_TEST:-postgres://splitplus_test:splitplus_test@localhost:${POSTGRES_PORT_TEST:-55432}/splitplus_test?sslmode=disable}"
    export DATABASE_URL_TEST
    cd be && go test -count=1 ./internal/integration/... -v
    @echo "Stopping integration test database..."
    docker compose -f docker/docker-compose.test.yml down -v --remove-orphans

# Run tests with coverage report
test-coverage:
    @echo "Running tests with coverage..."
    cd be && go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# ============================================================================
# Migrations (Run in Backend Container)
# ============================================================================

# Run database migrations up
migrate-up:
    @echo "Running database migrations..."
    docker compose -f docker/docker-compose.yml exec backend sh -c 'goose -dir internal/db/migrations postgres "$DATABASE_URL" up'

# Rollback last migration
migrate-down:
    @echo "Rolling back last migration..."
    docker compose -f docker/docker-compose.yml exec backend sh -c 'goose -dir internal/db/migrations postgres "$DATABASE_URL" down'

# Show migration status
migrate-status:
    @echo "Checking migration status..."
    docker compose -f docker/docker-compose.yml exec backend sh -c 'goose -dir internal/db/migrations postgres "$DATABASE_URL" status'

# Create a new migration file
migrate-create name:
    @echo "Creating new migration: {{name}}"
    docker compose -f docker/docker-compose.yml exec backend goose -dir internal/db/migrations create {{name}} sql

# ============================================================================
# SQLC
# ============================================================================

# Generate sqlc code from queries
sqlc-generate:
    @echo "Generating sqlc code..."
    cd be && sqlc generate

# ============================================================================
# Golang Commands (Local)
# ============================================================================

# Build API and Worker binaries
be-build:
    @echo "Building API..."
    cd be && go build -o bin/api ./cmd/api
    @echo "Building Worker..."
    cd be && go build -o bin/worker ./cmd/worker
    @echo "Build complete!"

# Download dependencies
go-deps:
    @echo "Downloading dependencies..."
    cd be && go mod download

# Tidy dependencies
go-tidy:
    @echo "Tidying dependencies..."
    cd be && go mod tidy

# Format Go code
be-fmt:
    @echo "Formatting Go code..."
    cd be && go fmt ./...

# Lint Go code
be-lint:
    @echo "Linting Go code..."
    cd be && go vet ./...

# ============================================================================
# TanStack Start Commands (Local - fe/)
# ============================================================================

# Start frontend dev server
fe-dev:
    @echo "Starting frontend dev server..."
    cd fe && npm run dev

# Build frontend for production
fe-build:
    @echo "Building frontend..."
    cd fe && npm run build

# Preview production build
fe-preview:
    @echo "Previewing production build..."
    cd fe && npm run preview

# Run frontend tests
fe-test:
    @echo "Running frontend tests..."
    cd fe && npm run test

# Run ESLint
fe-lint:
    @echo "Running ESLint..."
    cd fe && npm run lint

# Run Prettier
fe-format:
    @echo "Running Prettier..."
    cd fe && npm run format

# Format and lint
fe-check:
    @echo "Formatting and linting..."
    cd fe && npm run check
