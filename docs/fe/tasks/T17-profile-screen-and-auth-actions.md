# T17 Profile Screen and Auth Actions

## Depends On
- T03
- T04

## Goal
Implement profile and session actions.

## Scope
- Route: `/(app)/profile`.
- Load `GET /users/me`.
- Actions: `POST /auth/logout`, `POST /auth/logout-all`.

## Deliverables
- Profile screen with user details.
- Logout controls with confirmations.

## Acceptance Criteria
- Logout clears session and redirects to login.
- Logout-all also clears local session state.
