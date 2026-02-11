# T03 Auth and Session Module

## Depends On
- T02

## Goal
Build token/session management including refresh flow.

## Scope
- Implement `src/lib/api/auth.ts` and `src/lib/session/store.ts`.
- Support login, refresh, logout, logout-all.
- Persist and hydrate access/refresh tokens.

## Deliverables
- Session store API.
- Auth API module.
- Refresh helper callable by request pipeline.

## Acceptance Criteria
- Login stores tokens.
- Refresh updates access token.
- Logout clears session state and persisted tokens.
