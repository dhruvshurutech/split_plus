# API Documentation Coverage Audit

## Summary

**Total API Endpoints**: 56+  
**Documented in Bruno**: 56+ (100%)  
**Missing Documentation**: 0  
**Last Updated**: 2026-02-02

## Recent Updates (2026-02-02)

✅ **Authentication Updated**: All endpoints now use Bearer token authentication  
✅ **Expense Fields Added**: CategoryID and Tags fields documented  
✅ **Split Types Documented**: All split types (equal, fixed, percentage, shares, custom) with examples  
✅ **Folder Documentation**: Added comprehensive docs to all API folders  

## Authentication

All protected endpoints require JWT Bearer token authentication:

```
Authorization: Bearer {access_token}
```

### Getting a Token
1. Create a user: `POST /users`
2. Login: `POST /auth/login`  
3. Tokens are automatically saved to environment variables
4. Use `access_token` for all authenticated requests

### Token Lifetimes
- **Access Token**: 7 days
- **Refresh Token**: 30 days

### Public Endpoints (No Auth Required)
- `POST /users` - Create user
- `POST /auth/login` - Login
- `POST /auth/refresh` - Refresh token
- `GET /categories/presets` - Get category presets
- `GET /invitations/{token}` - View invitation details

## Endpoint Categories

## ✅ Fully Documented

### Auth (4/4) ✅

- POST /auth/login
- POST /auth/refresh
- POST /auth/logout
- POST /auth/logout-all

### Users (1/1) ✅

- POST /users

### Groups (4/4) ✅

- POST /groups
- GET /groups
- GET /groups/{group_id}/members
- POST /groups/{group_id}/invitations

### Expenses (6/6) ✅

- POST /groups/{group_id}/expenses
- GET /groups/{group_id}/expenses
- GET /groups/{group_id}/expenses/{expense_id}
- PUT /groups/{group_id}/expenses/{expense_id}
- DELETE /groups/{group_id}/expenses/{expense_id}
- GET /groups/{group_id}/expenses/search

### Balances (4/4) ✅

- GET /users/me/balances
- GET /groups/{group_id}/balances
- GET /groups/{group_id}/balances/{user_id}
- GET /groups/{group_id}/debts

### Settlements (5/5) ✅

- POST /groups/{group_id}/settlements
- GET /groups/{group_id}/settlements
- GET /groups/{group_id}/settlements/{settlement_id}
- PUT /groups/{group_id}/settlements/{settlement_id}
- DELETE /groups/{group_id}/settlements/{settlement_id}
- PATCH /groups/{group_id}/settlements/{settlement_id}/status

### Friends (11/11) ✅

- POST /friends/requests - Send friend request
- GET /friends - List friends
- GET /friends/requests/incoming - List incoming requests
- GET /friends/requests/outgoing - List outgoing requests
- POST /friends/requests/{id}/accept - Accept friend request
- POST /friends/requests/{id}/decline - Decline friend request
- DELETE /friends/{friend_id} - Remove friend
- POST /friends/{friend_id}/expenses - Create friend expense
- GET /friends/{friend_id}/expenses - List friend expenses
- POST /friends/{friend_id}/settlements - Create friend settlement
- GET /friends/{friend_id}/settlements - List friend settlements

### Expense Comments (4/4) ✅

- GET /groups/{group_id}/expenses/{expense_id}/comments - List comments
- POST /groups/{group_id}/expenses/{expense_id}/comments - Create comment
- PUT /groups/{group_id}/expenses/{expense_id}/comments/{comment_id} - Update comment
- DELETE /groups/{group_id}/expenses/{expense_id}/comments/{comment_id} - Delete comment

### Expense Categories (6/6) ✅

- GET /categories/presets - Get category presets (public)
- GET /groups/{group_id}/categories - List group categories
- POST /groups/{group_id}/categories - Create category
- POST /groups/{group_id}/categories/from-presets - Create from presets
- PUT /groups/{group_id}/categories/{id} - Update category
- DELETE /groups/{group_id}/categories/{id} - Delete category

### Recurring Expenses (6/6) ✅

- GET /groups/{group_id}/recurring-expenses - List recurring expenses
- POST /groups/{group_id}/recurring-expenses - Create recurring expense
- GET /groups/{group_id}/recurring-expenses/{id} - Get recurring expense
- PUT /groups/{group_id}/recurring-expenses/{id} - Update recurring expense
- DELETE /groups/{group_id}/recurring-expenses/{id} - Delete recurring expense
- POST /groups/{group_id}/recurring-expenses/{id}/generate - Generate expense

### Group Activities (2/2) ✅

- GET /groups/{group_id}/activity - List group activities
- GET /groups/{group_id}/expenses/{expense_id}/history - Get expense history

### Invitations (3/3) ✅

- GET /invitations/{token} - Get invitation details
- POST /invitations/{token}/accept - Accept invitation
- POST /invitations/{token}/join - Join via invitation

## ❌ Missing Documentation

None! All endpoints are fully documented. 🎉
