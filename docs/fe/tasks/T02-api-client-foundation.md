# T02 API Client Foundation

## Depends On
- None

## Goal
Create a typed API client and shared response/error handling for backend envelope.

## Scope
- Implement `src/lib/api/client.ts`.
- Parse envelope: `{ status, data, error }`.
- Centralize base URL and common headers.

## Deliverables
- Typed request helper.
- Error normalization helper.
- Base URL config from env.

## Acceptance Criteria
- All methods return typed data or typed error.
- Envelope parsing handles string and string[] error messages.
- Ready for auth wrapper extension.
