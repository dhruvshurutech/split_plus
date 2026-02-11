# Split+ Mobile-First FE Route Tree and Screen API Matrix

This document defines the TanStack Start route structure and maps each screen to backend API contracts.

API source of truth: `/Users/dhruvsaxena/Dev/dhruvsaxena1998/split_plus/http`

## Principles

- Mobile-first baseline: 360px width.
- All protected screens require valid access token.
- Use loader data for initial render and mutations for writes.
- Keep each route independently reloadable.

## Proposed Route Tree

```text
/
├── /login
├── /signup
├── /(app)
│   ├── /                          (Home / dashboard)
│   ├── /groups
│   ├── /groups/$groupId
│   ├── /groups/$groupId/expenses/new
│   ├── /groups/$groupId/expenses/$expenseId
│   ├── /groups/$groupId/settlements/new
│   ├── /friends
│   ├── /friends/$friendId
│   └── /profile
└── /invite/$token
```

## TanStack File-Based Route Skeleton

```text
fe/src/routes/__root.tsx
fe/src/routes/login.tsx
fe/src/routes/signup.tsx
fe/src/routes/invite.$token.tsx
fe/src/routes/_app.tsx
fe/src/routes/_app/index.tsx
fe/src/routes/_app/groups/index.tsx
fe/src/routes/_app/groups/$groupId/index.tsx
fe/src/routes/_app/groups/$groupId/expenses/new.tsx
fe/src/routes/_app/groups/$groupId/expenses/$expenseId.tsx
fe/src/routes/_app/groups/$groupId/settlements/new.tsx
fe/src/routes/_app/friends/index.tsx
fe/src/routes/_app/friends/$friendId.tsx
fe/src/routes/_app/profile.tsx
```

## Screen to API Matrix

| Route                                        | Screen Purpose               | Load APIs                                                                                                                                                                                                         | Mutation APIs                                                                                                                                                                                                                                                                                                | Notes                                                      |
| -------------------------------------------- | ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------- |
| `/login`                                     | Authenticate existing user   | None                                                                                                                                                                                                              | `POST /auth/login`                                                                                                                                                                                                                                                                                           | Save access + refresh tokens.                              |
| `/signup`                                    | Create account               | None                                                                                                                                                                                                              | `POST /users/` then `POST /auth/login`                                                                                                                                                                                                                                                                       | Auto-login after signup.                                   |
| `/invite/$token`                             | Join from invitation         | `GET /invitations/{token}`                                                                                                                                                                                        | `POST /invitations/{token}/join` or `POST /invitations/{token}/accept`                                                                                                                                                                                                                                       | Public route with optional auth.                           |
| `/(app)/`                                    | Home summary                 | `GET /users/me`, `GET /users/me/balances`, `GET /groups/`                                                                                                                                                         | None                                                                                                                                                                                                                                                                                                         | Keep payload compact for mobile.                           |
| `/(app)/groups`                              | Group list                   | `GET /groups/`                                                                                                                                                                                                    | `POST /groups/`                                                                                                                                                                                                                                                                                              | Create group in bottom sheet/modal on mobile.              |
| `/(app)/groups/$groupId`                     | Group overview               | `GET /groups/{group_id}/members`, `GET /groups/{group_id}/expenses/`, `GET /groups/{group_id}/balances`, `GET /groups/{group_id}/debts`, `GET /groups/{group_id}/settlements/`, `GET /groups/{group_id}/activity` | `POST /groups/{group_id}/invitations`                                                                                                                                                                                                                                                                        | Use tabbed sections for small screens.                     |
| `/(app)/groups/$groupId/expenses/new`        | Add expense                  | `GET /groups/{group_id}/members`, `GET /groups/{group_id}/categories`, `GET /categories/presets`                                                                                                                  | `POST /groups/{group_id}/expenses/`                                                                                                                                                                                                                                                                          | Support equal split first; extend split types iteratively. |
| `/(app)/groups/$groupId/expenses/$expenseId` | Expense detail               | `GET /groups/{group_id}/expenses/{expense_id}/`, `GET /groups/{group_id}/expenses/{expense_id}/comments`, `GET /groups/{group_id}/expenses/{expense_id}/history`                                                  | `PUT /groups/{group_id}/expenses/{expense_id}/`, `DELETE /groups/{group_id}/expenses/{expense_id}/`, `POST /groups/{group_id}/expenses/{expense_id}/comments`, `PUT /groups/{group_id}/expenses/{expense_id}/comments/{comment_id}`, `DELETE /groups/{group_id}/expenses/{expense_id}/comments/{comment_id}` | Optimistic comments UX recommended.                        |
| `/(app)/groups/$groupId/settlements/new`     | Create settlement            | `GET /groups/{group_id}/members`, `GET /groups/{group_id}/debts`                                                                                                                                                  | `POST /groups/{group_id}/settlements/`                                                                                                                                                                                                                                                                       | Prefill from simplified debts where possible.              |
| `/(app)/friends`                             | Friend network + requests    | `GET /friends/`, `GET /friends/requests/incoming`, `GET /friends/requests/outgoing`                                                                                                                               | `POST /friends/requests`, `POST /friends/requests/{id}/accept`, `POST /friends/requests/{id}/decline`, `DELETE /friends/{friend_id}`                                                                                                                                                                         | Action-heavy screen: prioritize thumb reach zones.         |
| `/(app)/friends/$friendId`                   | 1:1 ledger                   | `GET /friends/{friend_id}/expenses`, `GET /friends/{friend_id}/settlements`                                                                                                                                       | `POST /friends/{friend_id}/expenses`, `POST /friends/{friend_id}/settlements`                                                                                                                                                                                                                                | Reuse expense/settlement form components.                  |
| `/(app)/profile`                             | Account and session controls | `GET /users/me`                                                                                                                                                                                                   | `POST /auth/logout`, `POST /auth/logout-all`                                                                                                                                                                                                                                                                 | Add session clear + token reset in client state.           |

## Shared FE Modules

| Module                       | Responsibility                                  |
| ---------------------------- | ----------------------------------------------- |
| `src/lib/api/client.ts`      | typed `fetch`, base URL, error envelope parsing |
| `src/lib/api/auth.ts`        | login, refresh, logout, token persistence       |
| `src/lib/api/groups.ts`      | groups, invitations, members                    |
| `src/lib/api/expenses.ts`    | expenses, search, comments, categories          |
| `src/lib/api/balances.ts`    | balances, debts                                 |
| `src/lib/api/settlements.ts` | settlements and status updates                  |
| `src/lib/api/friends.ts`     | friendships, friend expenses/settlements        |
| `src/lib/session/store.ts`   | in-memory + storage token sync                  |
| `src/components/mobile/*`    | mobile-first layout primitives                  |

## Mobile Navigation Model

- Bottom nav items: Home, Groups, Friends, Profile.
- Group detail uses top segmented tabs: Expenses, Balances, Settlements, Activity.
- Primary write actions use sticky bottom CTA buttons.

## Non-Functional Requirements

- Touch target minimum 44x44.
- Skeleton loading states on all list screens.
- Retry CTA on network failure.
- Handle 401 globally with refresh flow.
- Route-level pending indicators for transitions.
