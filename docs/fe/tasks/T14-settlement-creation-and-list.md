# T14 Settlement Creation and List

## Depends On
- T02
- T08
- T13

## Goal
Add settlement workflow in group context.

## Scope
- List settlements with `GET /groups/{group_id}/settlements/`.
- Create settlement via `POST /groups/{group_id}/settlements/`.
- Optional status update path for follow-up iteration.

## Deliverables
- Settlement list section.
- Settlement create form route.

## Acceptance Criteria
- Create flow works and list updates.
- Uses group debts context for payer/payee guidance.
