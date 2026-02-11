# Split+ Backend API Contracts (Bruno)

This directory contains a Bruno collection for all backend HTTP routes mounted in `/be/internal/app/app.go`.

## Collection Setup

- Open this folder in Bruno.
- Use `/Users/dhruvsaxena/Dev/dhruvsaxena1998/split_plus/http/environments/local.bru` and set:
  - `baseUrl` (default `http://localhost:8080`)
  - `accessToken` and `refreshToken` after login.
  - Entity IDs (`groupId`, `expenseId`, etc.) based on your test data.

## Response Envelope

All endpoints use a shared envelope from `/be/internal/http/response/response.go`:

- Success: `{ "status": true, "data": ... }`
- Error: `{ "status": false, "error": { "message": "..." } }`
- Validation errors may return `error.message` as `string[]`.

## Route Coverage

Total routes documented: **58**.

### Public Routes

- `POST /auth/login`
- `POST /auth/refresh`
- `POST /users/`
- `GET /categories/presets`
- `GET /invitations/{token}`

### Optional Auth Route

- `POST /invitations/{token}/join` (supports anonymous join flow, and also authenticated join)

### Protected Routes

All other requests in this collection require `Authorization: Bearer {{accessToken}}`.

## Contract Sources

Contracts were derived from:

- Route registration: `/be/internal/http/router/*.go`
- Request DTOs and param parsing: `/be/internal/http/handlers/*.go`
- Auth middleware: `/be/internal/http/middleware/auth.go`

Each `.bru` request includes route-level docs with:

- auth requirements
- path/query params
- request body contract
- response envelope notes
