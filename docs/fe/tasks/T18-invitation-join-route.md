# T18 Invitation Join Route

## Depends On
- T02

## Goal
Build invitation entry and join/accept flow.

## Scope
- Route: `/invite/$token`.
- Load invitation details.
- For guest: join endpoint with optional name/password.
- For logged-in user: accept endpoint path.

## Deliverables
- Invitation detail screen.
- Conditional action logic (guest vs authenticated).

## Acceptance Criteria
- Valid token shows invitation data.
- Join/accept success routes user into app flow.
