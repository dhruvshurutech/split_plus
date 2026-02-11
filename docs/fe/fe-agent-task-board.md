# Split+ FE Agent Task Board (Dependency Aware)

This board is structured so multiple FE agents can work in parallel with clear dependency gates.

## Dependency Legend

- `None`: can start immediately.
- `T#`: depends on completion of task id.

## Task List

| ID  | Task                              | Deliverables                                                                          | Depends On                  | Parallelizable |
| --- | --------------------------------- | ------------------------------------------------------------------------------------- | --------------------------- | -------------- |
| T1  | App shell and route scaffolding   | file routes created per route tree, mobile app shell layout, bottom nav               | None                        | Yes            |
| T2  | API client foundation             | `src/lib/api/client.ts`, response envelope parser, typed error model, base URL config | None                        | Yes            |
| T3  | Auth and session module           | token store, login/logout/refresh services, auth guard utility                        | T2                          | No             |
| T4  | Protected route guard integration | `_app` route guard using session state, redirect logic                                | T1, T3                      | No             |
| T5  | Login and signup screens          | `/login`, `/signup` pages, form validation, success redirects                         | T1, T3                      | Yes            |
| T6  | Home dashboard                    | `/` screen with me + balances + groups summary cards                                  | T4, T2                      | Yes            |
| T7  | Groups list and create group      | `/groups` UI with list + create modal/sheet                                           | T4, T2                      | Yes            |
| T8  | Group detail tabs scaffold        | `/groups/$groupId` tabs shell: Expenses/Balances/Settlements/Activity                 | T4, T1                      | Yes            |
| T9  | Group members and invitations     | members list and invite flow in group detail                                          | T8, T2                      | Yes            |
| T10 | Expense creation flow             | `/groups/$groupId/expenses/new` form (equal split first), submit/create               | T8, T2                      | Yes            |
| T11 | Expense list and detail           | list in group tab, detail page with update/delete                                     | T8, T2                      | Yes            |
| T12 | Comments and expense history      | comments CRUD + history rendering on expense detail                                   | T11, T2                     | No             |
| T13 | Balances and debts tab            | group balances and simplified debts cards                                             | T8, T2                      | Yes            |
| T14 | Settlement creation and list      | settlement tab list + create screen                                                   | T8, T13, T2                 | No             |
| T15 | Friends hub                       | `/friends` list, incoming/outgoing requests, request actions                          | T4, T2                      | Yes            |
| T16 | Friend ledger page                | `/friends/$friendId` expenses + settlements + create actions                          | T15, T2                     | No             |
| T17 | Profile screen and auth actions   | `/profile`, me data, logout/logout-all actions                                        | T4, T3                      | Yes            |
| T18 | Invitation join route             | `/invite/$token` token lookup + join/accept flow                                      | T2                          | Yes            |
| T19 | Global UX hardening               | loading skeletons, empty states, error states, retry CTA                              | T6, T7, T8, T11, T13, T15   | Yes            |
| T20 | Test coverage and CI checks       | tests for auth refresh, guards, key forms/screens, build/lint green                   | T5, T10, T11, T14, T15, T17 | No             |

## Suggested Parallel Execution Waves

| Wave   | Tasks                    |
| ------ | ------------------------ |
| Wave 1 | T1, T2                   |
| Wave 2 | T3, T5, T18              |
| Wave 3 | T4, T6, T7, T8, T15, T17 |
| Wave 4 | T9, T10, T11, T13, T16   |
| Wave 5 | T12, T14, T19            |
| Wave 6 | T20                      |

## Agent Assignment Example

| Agent      | Initial Picks           |
| ---------- | ----------------------- |
| FE-Agent-A | T1 -> T4 -> T8          |
| FE-Agent-B | T2 -> T3 -> T5          |
| FE-Agent-C | T7 -> T9                |
| FE-Agent-D | T15 -> T16              |
| FE-Agent-E | T10 -> T11 -> T12       |
| FE-Agent-F | T13 -> T14              |
| FE-Agent-G | T6 -> T17 -> T19 -> T20 |

## Task Completion Criteria

Each task is considered complete only when:

- matching APIs are wired from `/http` Bruno contracts,
- mobile layout verified at 360px width,
- loading and error states are implemented,
- lint and relevant tests pass.
