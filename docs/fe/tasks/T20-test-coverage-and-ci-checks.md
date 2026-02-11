# T20 Test Coverage and CI Checks

## Depends On
- T05
- T10
- T11
- T14
- T15
- T17

## Goal
Add confidence coverage and ensure FE pipeline readiness.

## Scope
- Unit tests for auth/session and API client paths.
- Component/integration tests for critical forms and flows.
- Ensure lint, test, and build pass.

## Deliverables
- Test files for high-risk paths.
- Green checks for `lint`, `test`, and `build`.

## Acceptance Criteria
- Critical user flows covered: auth, group create, expense create, settlements, friend actions.
- No failing FE checks in CI-equivalent local run.
