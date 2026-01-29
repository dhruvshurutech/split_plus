# Split+ Project Overview

## Purpose
Split+ is a Golang-based backend for a bill splitting application.

## Tech Stack
- **Language**: Go
- **Database**: PostgreSQL with `pgx` and `sqlc` for type-safe queries.
- **Migrations**: `goose`.
- **Router**: `chi`.
- **Validation**: `go-playground/validator`.
- **Docker**: Used for development (application and postgres).

## Project Structure
- `cmd/api/`: Application entry point.
- `internal/app/`: Application initialization and dependency injection.
- `internal/db/`: Database migrations (`migrations/`), sqlc configuration, and queries (`queries/`).
- `internal/http/`: Handlers, middleware, and router.
- `internal/repository/`: Data access layer.
- `internal/service/`: Business logic layer.
- `bruno/`: API documentation and collections for the Bruno API client.

## Authentication
Authentication is currently implemented via a simple `X-User-ID` header check in the `RequireAuth` middleware.
