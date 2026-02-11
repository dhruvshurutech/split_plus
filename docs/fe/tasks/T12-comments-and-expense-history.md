# T12 Comments and Expense History

## Depends On
- T02
- T11

## Goal
Add collaborative context to expense detail.

## Scope
- Comments: list/create/update/delete APIs.
- Expense history: `GET /groups/{group_id}/expenses/{expense_id}/history`.

## Deliverables
- Comments thread UI.
- History feed section.

## Acceptance Criteria
- Full comment CRUD works.
- History entries render with sensible formatting.
