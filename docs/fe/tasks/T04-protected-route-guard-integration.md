# T04 Protected Route Guard Integration

## Depends On
- T01
- T03

## Goal
Protect authenticated routes and redirect unauthenticated users.

## Scope
- Add guard logic to `_app` route boundary.
- Redirect unauthenticated users to `/login`.
- Preserve return path for post-login redirect.

## Deliverables
- Guard utility integration with router.
- Redirect behavior for protected routes.

## Acceptance Criteria
- Protected routes cannot be accessed without session.
- Authenticated users remain in app area.
