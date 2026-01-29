# Code Style and Conventions

## Naming
- Follow standard Go naming conventions (camelCase for internal, PascalCase for exported).
- Interfaces and options are used for dependency injection and configuration.

## Design Patterns
- **Repository Pattern**: Data access is abstracted in the `repository` package.
- **Service Layer**: Business logic is separated into the `service` package.
- **Dependency Injection**: Dependencies are passed into constructors (e.g., `NewUserService(repo)`).
- **Middleware Options**: The router uses a functional options pattern (`Option`, `optionFunc`) for configuration.

## Formatting
- Use `go fmt` (available via `just fmt`).
- Use `go vet` for static analysis (available via `just vet`).
