# Split+ Frontend Integration Guide

This document provides the API flow and implementation details for integrating the Split+ backend with the frontend application.

## 1. Authentication & API Basics

### Base URL

`http://localhost:8080` (Development)

### Authentication

Currently, the API uses a header-based authentication for simplicity during development. Every protected request must include the `X-User-ID` header.

- **Header Name:** `X-User-ID`
- **Value:** User's UUID (e.g., `550e8400-e29b-41d4-a716-446655440000`)

> **Note:** Actual JWT authentication is planned for future phases.

---

## 2. API Reference

### 👤 Users & Onboarding

| Method | Endpoint | Description               | Auth Required |
| :----- | :------- | :------------------------ | :------------ |
| `POST` | `/users` | Create a new user account | No            |

### 👥 Groups

| Method | Endpoint                         | Description                                            | Auth Required |
| :----- | :------------------------------- | :----------------------------------------------------- | :------------ |
| `GET`  | `/groups`                        | List all groups the user belongs to                    | Yes           |
| `POST` | `/groups`                        | Create a new group                                     | Yes           |
| `GET`  | `/groups/{group_id}/members`     | List all members of a specific group                   | Yes           |
| `POST` | `/groups/{group_id}/invitations` | Invite a user to the group via email                   | Yes           |
| `GET`  | `/invitations/{token}`           | Get invitation details (public info)                   | No            |
| `POST` | `/invitations/{token}/accept`    | Accept an invitation and join group                    | Yes           |
| `POST` | `/invitations/{token}/join`      | Smart Join (handles login/register/accept in one flow) | No            |

### 🧾 Expenses & Categories

| Method   | Endpoint                                    | Description                             | Auth Required |
| :------- | :------------------------------------------ | :-------------------------------------- | :------------ |
| `GET`    | `/groups/{group_id}/expenses`               | List group expenses (paginated)         | Yes           |
| `GET`    | `/groups/{group_id}/expenses/search`        | Advanced search for expenses            | Yes           |
| `POST`   | `/groups/{group_id}/expenses`               | Create a new group expense              | Yes           |
| `GET`    | `/groups/{group_id}/expenses/{id}`          | Get detailed expense (including splits) | Yes           |
| `PUT`    | `/groups/{group_id}/expenses/{id}`          | Update an existing expense              | Yes           |
| `DELETE` | `/groups/{group_id}/expenses/{id}`          | Delete an expense                       | Yes           |
| `GET`    | `/categories/presets`                       | Get system category presets             | No            |
| `GET`    | `/groups/{group_id}/categories`             | List categories available in a group    | Yes           |
| `GET`    | `/groups/{group_id}/expenses/{id}/comments` | List comments on an expense             | Yes           |
| `POST`   | `/groups/{group_id}/expenses/{id}/comments` | Add a comment to an expense             | Yes           |

### 🔄 Recurring Expenses

| Method   | Endpoint                                              | Description                          | Auth Required |
| :------- | :---------------------------------------------------- | :----------------------------------- | :------------ |
| `GET`    | `/groups/{group_id}/recurring-expenses`               | List all recurring expense templates | Yes           |
| `POST`   | `/groups/{group_id}/recurring-expenses`               | Create a new recurring template      | Yes           |
| `DELETE` | `/groups/{group_id}/recurring-expenses/{id}`          | Stop/Delete a recurring expense      | Yes           |
| `POST`   | `/groups/{group_id}/recurring-expenses/{id}/generate` | Manually trigger expense generation  | Yes           |

### 💰 Balances & Settlements

| Method | Endpoint                         | Description                                | Auth Required |
| :----- | :------------------------------- | :----------------------------------------- | :------------ |
| `GET`  | `/users/me/balances`             | Get overall user balance across all groups | Yes           |
| `GET`  | `/groups/{group_id}/balances`    | List all member balances in a group        | Yes           |
| `GET`  | `/groups/{group_id}/debts`       | Get "Who owes whom" simplified view        | Yes           |
| `POST` | `/groups/{group_id}/settlements` | Record a payment to settle a debt          | Yes           |

### 🤝 Friends

| Method | Endpoint                        | Description                        | Auth Required |
| :----- | :------------------------------ | :--------------------------------- | :------------ |
| `GET`  | `/friends`                      | List all friends                   | Yes           |
| `POST` | `/friends/requests`             | Send a friend request              | Yes           |
| `POST` | `/friends/{friend_id}/expenses` | Create a 1:1 expense with a friend | Yes           |

### 📈 Activity & History

| Method | Endpoint                                   | Description                          | Auth Required |
| :----- | :----------------------------------------- | :----------------------------------- | :------------ |
| `GET`  | `/groups/{group_id}/activity`              | Get group activity feed              | Yes           |
| `GET`  | `/groups/{group_id}/expenses/{id}/history` | Get audit log for a specific expense | Yes           |

---

## 3. User Flows & Page Mapping

### Home / Dashboard

- **Initial Load:**
  1.  `GET /users/me/balances` -> Display "You are owed $X" and "You owe $Y".
  2.  `GET /groups` -> Display list of active groups.
  3.  `GET /friends` -> Display active contacts.

### Group Details Page

- **Page Load:**
  1.  `GET /groups/{group_id}/expenses` -> Recent activity feed.
  2.  `GET /groups/{group_id}/balances` -> Sidebar showing member status.
  3.  `GET /groups/{group_id}/debts` -> "Settle up" suggestions.
- **Actions:**
  - **Add Expense:** Requires `GET /groups/{group_id}/categories` to show category picker.
  - **Invite Member:** `POST /groups/{group_id}/invitations`.

### Individual Expense View

- **Page Load:**
  1.  `GET /groups/{group_id}/expenses/{expense_id}` -> Full split details.
  2.  `GET /groups/{group_id}/expenses/{expense_id}/comments` -> Discussion thread.
  3.  `GET /groups/{group_id}/expenses/{expense_id}/history` -> Audit log of changes.

---

## 4. Detailed Implementation Flows

### Flow: Creating a Group Expense

This is the most complex flow and requires careful payload construction. It supports splitting with both registered users and invited (pending) users.

1.  **Preparation**:
    - Fetch `GET /groups/{group_id}/members` to get the list of members. Note that some members might have a `user_id` while others (pending) only have a `pending_user_id`.
    - Fetch `GET /groups/{group_id}/categories` for classification.
2.  **Input**: User enters title, amount, date, selects payer, and selects split method.
3.  **Submission**:
    - **Endpoint**: `POST /groups/{group_id}/expenses`
    - **Payload Example**:
      ```json
      {
        "title": "Grocery Shopping",
        "amount": "60.00",
        "date": "2026-01-28",
        "category_id": "uuid-here",
        "payments": [{ "user_id": "payer-uuid", "amount": "60.00" }],
        "splits": [
          { "user_id": "user1-uuid", "amount_owned": "30.00" },
          { "pending_user_id": "invited-user-uuid", "amount_owned": "30.00" }
        ]
      }
      ```
    - **Note**: Use either `user_id` or `pending_user_id` for each entry in `payments` and `splits`.

### Flow: Invitations and Member Management

1.  **Invite by Email**: `POST /groups/{group_id}/invitations`. If the email isn't registered, a "pending user" is created.
2.  **Member List**: `GET /groups/{group_id}/members` returns all members.
    - **Accepted Members**: Have a `user_id`.
    - **Pending Members**: Have a `pending_user_id` and `status: "invited"` (or similar).
    - **Frontend Logic**: When creating expenses, you can use either ID type. The backend handles the merge once they join.

### Flow: Accepting an Invitation (Smart Join)

This is the recommended flow for both new and existing users. It minimizes friction by combining authentication and joining.

1.  **User lands on URL**: `/join/{token}`
2.  **Page Load**: `GET /invitations/{token}` to show group name and inviter.
3.  **Authentication Check**:
    - **If already logged in**: Show "Join Group" button.
    - **If not logged in**: Show Password field (and optional Name field).
4.  **Action**: User clicks "Join Group".
5.  **Submission**: `POST /invitations/{token}/join`
    - **Headers**: Include `X-User-ID` if logged in.
    - **Payload**: `{ "password": "...", "name": "..." }` (only needed if NOT logged in).
6.  **Backend Magic**:
    - **Scenario A (Authenticated)**: If `X-User-ID` is provided, the backend verifies the email matches the invitation and joins immediately.
    - **Scenario B (Existing Account)**: If the email is registered but no header is sent, the backend validates the `password` and joins.
    - **Scenario C (New Account)**: If the email is unknown, the backend creates a new account with the provided `password` and joins.
    - **Global Data Merge**: In ALL scenarios, the backend automatically finds any expenses previously assigned to that email (as a "pending user") and transfers them to the real account across **ALL groups**.

### Flow: Settlements (Settling Up)

1.  **View Debts**: `GET /groups/{group_id}/debts` shows who needs to pay whom.
2.  **Record Settlement**: `POST /groups/{group_id}/settlements`.
    - **Payload**:
      ```json
      {
        "from_user_id": "uuid",
        "to_user_id": "uuid",
        "amount": "15.50",
        "currency_code": "USD",
        "date": "2026-01-28",
        "payment_method": "cash"
      }
      ```

---

## 5. Error Handling

The API returns consistent error responses:

```json
{
  "status": "error",
  "message": "Detailed error message here"
}
```

### Common Status Codes

- **400 Bad Request**: Invalid input or missing fields.
- **401 Unauthorized**: Missing or invalid `X-User-ID` header.
- **403 Forbidden**: User is not a member of the group or lacks permissions.
- **404 Not Found**: Resource (group, expense, etc.) doesn't exist.
- **409 Conflict**: Duplicate data (e.g., user already in group).
- **422 Unprocessable Entity**: Business logic error (e.g., splits don't sum up to total amount).

---

## 6. Recommendations for Frontend Team

1.  **Caching**: Use a library like `react-query` or `swr` to cache group members and categories, as they don't change frequently.
2.  **UUID Handling**: Ensure all IDs are treated as strings.
3.  **Date Format**: Always use `YYYY-MM-DD` for request payloads.
4.  **Currency**: The backend handles amounts as strings to maintain precision. Avoid converting to floats in the frontend where possible; use libraries like `big.js` or `dinero.js` for calculations.
