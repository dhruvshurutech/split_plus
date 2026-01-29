# Suggested Commands

## Running the Project
- `just up`: Start all services (app + postgres) in Docker.
- `just down`: Stop all services.
- `just run`: Run the application locally (requires `DATABASE_URL`).
- `just setup`: Start services, run migrations, and generate sqlc code.

## Development & Maintenance
- `just sqlc-generate`: Generate sqlc code from queries.
- `just migrate-up`: Run database migrations.
- `just migrate-create <name>`: Create a new migration file.
- `just fmt`: Format Go code.
- `just vet`: Run go vet.
- `just tidy`: Tidy dependencies.

## Testing
- `just test`: Run all tests.
- `just test-verbose`: Run tests with verbose output.
- `just test-coverage`: Run tests and generate HTML coverage report.
- `just test-coverage-summary`: Run tests and show coverage summary in terminal.

## Utilities
- `just shell`: Open a shell in the app container.
- `just db-connect`: Connect to the PostgreSQL database using psql.
- `just logs`: View logs from all services.
