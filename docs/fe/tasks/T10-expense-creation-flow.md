# T10 Expense Creation Flow

## Depends On
- T02
- T08

## Goal
Create mobile-first add-expense flow for groups.

## Scope
- Route: `/(app)/groups/$groupId/expenses/new`.
- Load members and categories.
- Submit `POST /groups/{group_id}/expenses/`.
- Support equal split first.

## Deliverables
- Expense create form.
- Payment/split editors.

## Acceptance Criteria
- Valid expense creates successfully.
- Field and server validation errors handled.
