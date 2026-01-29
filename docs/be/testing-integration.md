# Integration testing (E2E) with real Postgres + `httptest.Server`

This repo includes an **automated end-to-end integration suite** that exercises the full flow:
**users → groups → expenses → balances → settlements → recurring (manual generate)**.

It runs via `go test` using `httptest.Server` (no external API process), but uses a **real Postgres test database**.

## Prerequisites

- Docker (for the test Postgres container)
- Go toolchain

## 1) Start the test database

```bash
just test-db-up
```

The test DB container is defined in `../../docker/docker-compose.test.yml` and maps Postgres to:
- default host port: `55432` (override via `POSTGRES_PORT_TEST`)

## 2) Set `DATABASE_URL_TEST`

You can export it (defaults shown):

```bash
export DATABASE_URL_TEST="postgres://splitplus_test:splitplus_test@localhost:55432/splitplus_test?sslmode=disable"
```

You may override any of these env vars:
- `DATABASE_URL_TEST`
- `POSTGRES_PORT_TEST` (default `55432`)
- `POSTGRES_USER` (default `splitplus_test`)
- `POSTGRES_PASSWORD` (default `splitplus_test`)
- `POSTGRES_DB` (default `splitplus_test`)

## 3) Run the integration suite

```bash
just test-integration
```

This does:
- ensure the test DB is up
- run `go test -count=1 ./internal/integration/...`

## What the suite does

The harness resets DB state per run:
- drops and recreates `public` schema
- enables `pgcrypto` (for `gen_random_uuid()`)
- applies all migrations from `internal/db/migrations`

Then `TestFullFlow_HappyPath` performs the end-to-end flow by calling the real HTTP handlers.

## Troubleshooting

- **`DATABASE_URL_TEST is not set`**
  - Export `DATABASE_URL_TEST` (see above).
- **Migration failures mentioning `gen_random_uuid`**
  - Ensure Postgres is running; the harness enables `pgcrypto` automatically.
- **Port already in use**
  - Override `POSTGRES_PORT_TEST`, e.g. `POSTGRES_PORT_TEST=55433 just test-integration`.

